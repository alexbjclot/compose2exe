package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const maxEmbedSize = 2 * 1024 * 1024 * 1024 // 2 GB Go embed limit

type Generator struct {
	ComposePath string
	Name        string
	Embed       bool
	Output      string
	Targets     []string
	Module      string
}

type FileVolume struct {
	OriginalVolume string
	OriginalName   string
	SafeName       string
	MountPath      string
	HostPath       string
}

func (g *Generator) Run() error {
	compose, err := ParseCompose(g.ComposePath)
	if err != nil {
		return err
	}

	composeDir := filepath.Dir(g.ComposePath)

	// Step 1: Build images for services that use build:
	for name, svc := range compose.Services {
		if svc.Build != "" && svc.Image == "" {
			imageName := fmt.Sprintf("%s:latest", sanitize(name))
			buildContext := filepath.Join(composeDir, svc.Build)
			fmt.Printf("docker build -t %s %s\n", imageName, buildContext)
			cmd := exec.Command("docker", "build", "-t", imageName, buildContext)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("docker build %s: %w", name, err)
			}
			svc.Image = imageName
			svc.AutoEmbed = true // mark for automatic embed
		}
	}

	// Step 2: Detect file volumes
	fileVolumes := detectFileVolumes(compose, composeDir)

	// Step 3: Create build dir
	buildDir, err := os.MkdirTemp("", "compose2exe-*")
	if err != nil {
		return fmt.Errorf("cannot create build dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	// Step 4: Init Go module
	fmt.Printf("go mod init %s\n", g.Module)
	if err := runCmd(buildDir, "go", "mod", "init", g.Module); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}

	// Step 5: Embed images
	// - Always embed services with build: (AutoEmbed)
	// - Embed all services if --embed flag is set
	// - Check size limit before embedding
	embedServices := make(map[string]bool)
	for name, svc := range compose.Services {
		if svc.AutoEmbed || g.Embed {
			if svc.Image == "" {
				continue
			}
			// Check image size before embedding
			size, err := getImageSize(svc.Image)
			if err != nil {
				fmt.Printf("Warning: cannot check size of %s: %v\n", svc.Image, err)
			}
			if size > maxEmbedSize {
				if svc.AutoEmbed {
					return fmt.Errorf(
						"service '%s' uses build: but the resulting image is too large to embed (%.1f GB > 2 GB limit).\n"+
							"  → Push the image to a registry and use image: instead of build:",
						name, float64(size)/1024/1024/1024,
					)
				}
				fmt.Printf("Warning: skipping embed for %s (%.1f GB > 2 GB limit)\n", name, float64(size)/1024/1024/1024)
				continue
			}
			tarName := fmt.Sprintf("%s.tar.gz", sanitize(name))
			fmt.Printf("docker save %s → %s\n", svc.Image, tarName)
			if err := runDockerSave(svc.Image, filepath.Join(buildDir, tarName)); err != nil {
				return fmt.Errorf("docker save %s: %w", svc.Image, err)
			}
			embedServices[name] = true
		}
	}

	// Step 6: Copy file volumes
	if len(fileVolumes) > 0 {
		volDir := filepath.Join(buildDir, "volumes")
		if err := os.MkdirAll(volDir, 0755); err != nil {
			return err
		}
		for _, fv := range fileVolumes {
			src := filepath.Join(composeDir, fv.HostPath)
			dst := filepath.Join(volDir, fv.SafeName)
			fmt.Printf("Embedding file volume: %s\n", fv.OriginalName)
			if err := copyFile(src, dst); err != nil {
				fmt.Printf("  (file not found, creating empty: %s)\n", fv.OriginalName)
				if err2 := os.WriteFile(dst, []byte{}, 0644); err2 != nil {
					return err2
				}
			}
		}
	}

	// Step 7: Collect ports for preflight
	allPorts := collectPorts(compose)

	// Step 8: Separate plain volumes from file volumes per service
	for _, svc := range compose.Services {
		svc.PlainVolumes = filterPlainVolumes(svc.Volumes, fileVolumes)
	}

	// Step 9: Generate main.go
	mainGo, err := g.renderMain(compose, fileVolumes, allPorts, embedServices)
	if err != nil {
		return fmt.Errorf("render main.go: %w", err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	// Step 10: go get + build
	fmt.Println("go get")
	runCmd(buildDir, "go", "get")

	if err := os.MkdirAll(g.Output, 0755); err != nil {
		return err
	}

	for _, target := range g.Targets {
		parts := strings.Split(target, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid target: %s", target)
		}
		goos, goarch := parts[0], parts[1]
		outName := fmt.Sprintf("%s-%s-%s", g.Name, goos, goarch)
		if goos == "windows" {
			outName += ".exe"
		}
		outPath := filepath.Join(g.Output, outName)
		fmt.Printf("GOOS=%s GOARCH=%s go build -o %s\n", goos, goarch, outPath)
		cmd := exec.Command("go", "build", "-o", outPath)
		cmd.Dir = buildDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOOS=%s", goos), fmt.Sprintf("GOARCH=%s", goarch))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build %s: %w", target, err)
		}
	}

	return nil
}

func (g *Generator) renderMain(compose *ComposeFile, fileVolumes []FileVolume, allPorts []string, embedServices map[string]bool) (string, error) {
	tmpl, err := template.New("main").Funcs(template.FuncMap{
		"sanitize": sanitize,
	}).Parse(mainTemplate)
	if err != nil {
		return "", err
	}

	order := GetOrderedServices(compose)
	hasEmbed := len(embedServices) > 0

	data := struct {
		Compose        *ComposeFile
		Order          []string
		Embed          bool
		HasEmbed       bool
		EmbedServices  map[string]bool
		Networks       map[string]*Network
		FileVolumes    []FileVolume
		HasFileVolumes bool
		AllPorts       []string
	}{
		Compose:        compose,
		Order:          order,
		Embed:          g.Embed,
		HasEmbed:       hasEmbed,
		EmbedServices:  embedServices,
		Networks:       compose.Networks,
		FileVolumes:    fileVolumes,
		HasFileVolumes: len(fileVolumes) > 0,
		AllPorts:       allPorts,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// getImageSize returns the size of a Docker image in bytes
func getImageSize(image string) (int64, error) {
	out, err := exec.Command("docker", "inspect", "--format={{.Size}}", image).Output()
	if err != nil {
		return 0, err
	}
	var size int64
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &size)
	return size, nil
}

func detectFileVolumes(compose *ComposeFile, composeDir string) []FileVolume {
	var result []FileVolume
	seen := make(map[string]bool)

	for _, svc := range compose.Services {
		for _, vol := range svc.Volumes {
			parts := strings.SplitN(vol, ":", 2)
			if len(parts) != 2 {
				continue
			}
			hostPath := parts[0]
			mountPath := parts[1]

			if !strings.HasPrefix(hostPath, "./") && !strings.HasPrefix(hostPath, "/") && !strings.HasPrefix(hostPath, "../") {
				continue
			}

			absPath := filepath.Join(composeDir, hostPath)
			info, err := os.Stat(absPath)
			isFile := err != nil || !info.IsDir()

			if isFile && !seen[vol] {
				seen[vol] = true
				name := filepath.Base(hostPath)
				result = append(result, FileVolume{
					OriginalVolume: vol,
					OriginalName:   name,
					SafeName:       sanitize(name),
					MountPath:      mountPath,
					HostPath:       hostPath,
				})
			}
		}
	}
	return result
}

func filterPlainVolumes(volumes []string, fileVolumes []FileVolume) []string {
	fileVolMap := make(map[string]bool)
	for _, fv := range fileVolumes {
		fileVolMap[fv.OriginalVolume] = true
	}
	var plain []string
	for _, v := range volumes {
		if !fileVolMap[v] {
			plain = append(plain, v)
		}
	}
	return plain
}

func collectPorts(compose *ComposeFile) []string {
	seen := make(map[string]bool)
	var ports []string
	for _, svc := range compose.Services {
		for _, p := range svc.Ports {
			parts := strings.Split(p, ":")
			host := parts[0]
			if !seen[host] {
				seen[host] = true
				ports = append(ports, host)
			}
		}
	}
	return ports
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerSave(image, tarPath string) error {
	save := exec.Command("docker", "save", image)
	gzip := exec.Command("gzip")
	outFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	gzip.Stdin, _ = save.StdoutPipe()
	gzip.Stdout = outFile
	gzip.Stderr = os.Stderr
	save.Stderr = os.Stderr
	if err := gzip.Start(); err != nil {
		return err
	}
	if err := save.Run(); err != nil {
		return err
	}
	return gzip.Wait()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func sanitize(s string) string {
	return strings.NewReplacer("-", "_", ".", "_", "/", "_", ":", "_", " ", "_").Replace(s)
}
