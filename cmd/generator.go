package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator holds the configuration for generating the executable
type Generator struct {
	ComposePath string
	Name        string
	Embed       bool
	Output      string
	Targets     []string
	Module      string
}

// Run parses the compose file and generates the executable
func (g *Generator) Run() error {
	compose, err := ParseCompose(g.ComposePath)
	if err != nil {
		return err
	}

	// Create a temp build dir
	buildDir, err := os.MkdirTemp("", "compose2exe-*")
	if err != nil {
		return fmt.Errorf("cannot create build dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	// Init Go module
	fmt.Printf("go mod init %s\n", g.Module)
	if out, err := runCmd(buildDir, "go", "mod", "init", g.Module); err != nil {
		return fmt.Errorf("go mod init: %s\n%w", out, err)
	}

	// If embed, save each image as a tarball
	if g.Embed {
		for name, svc := range compose.Services {
			if svc.Image == "" {
				fmt.Printf("Skipping service %s (no image)\n", name)
				continue
			}
			tarName := fmt.Sprintf("%s.tar.gz", sanitize(name))
			fmt.Printf("docker save %s → %s\n", svc.Image, tarName)
			tarPath := filepath.Join(buildDir, tarName)
			if out, err := runDockerSave(svc.Image, tarPath); err != nil {
				return fmt.Errorf("docker save %s: %s\n%w", svc.Image, out, err)
			}
		}
	}

	// Generate main.go from template
	mainGo, err := g.renderMain(compose)
	if err != nil {
		return fmt.Errorf("render main.go: %w", err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	// go get dependencies
	fmt.Println("go get")
	if out, err := runCmd(buildDir, "go", "get"); err != nil {
		return fmt.Errorf("go get: %s\n%w", out, err)
	}

	// Build for each target
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

func (g *Generator) renderMain(compose *ComposeFile) (string, error) {
	tmpl, err := template.New("main").Funcs(template.FuncMap{
		"sanitize": sanitize,
		"join":     strings.Join,
	}).Parse(mainTemplate)
	if err != nil {
		return "", err
	}

	order := GetOrderedServices(compose)

	data := struct {
		Compose  *ComposeFile
		Order    []string
		Embed    bool
		Module   string
		Networks map[string]*Network
	}{
		Compose:  compose,
		Order:    order,
		Embed:    g.Embed,
		Module:   g.Module,
		Networks: compose.Networks,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func runCmd(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return "", cmd.Run()
}

func runDockerSave(image, tarPath string) (string, error) {
	// docker save image | gzip > tarPath
	save := exec.Command("docker", "save", image)
	gzip := exec.Command("gzip")
	outFile, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	gzip.Stdin, _ = save.StdoutPipe()
	gzip.Stdout = outFile
	gzip.Stderr = os.Stderr
	save.Stderr = os.Stderr

	if err := gzip.Start(); err != nil {
		return "", err
	}
	if err := save.Run(); err != nil {
		return "", err
	}
	return "", gzip.Wait()
}

func sanitize(s string) string {
	return strings.NewReplacer("-", "_", ".", "_", "/", "_", ":", "_").Replace(s)
}
