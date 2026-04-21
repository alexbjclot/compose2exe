package cmd

const mainTemplate = `package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
{{- if .HasFileVolumes}}
	"path/filepath"
{{- end}}
	"runtime"
	"strings"
	"syscall"
{{- if .Embed}}
	"bytes"
	_ "embed"
{{- end}}
{{- if .HasFileVolumes}}
	_ "embed"
{{- end}}
)

{{- if .Embed}}
{{- range .Order}}{{$sn := sanitize .}}{{with index $.Compose.Services .}}{{if .Image}}
//go:embed {{$sn}}.tar.gz
var embedded_{{$sn}} []byte
{{- end}}{{end}}{{end}}
{{- end}}

{{- range .FileVolumes}}
//go:embed volumes/{{.SafeName}}
var fileVolume_{{.SafeName}} []byte
{{- end}}

var containerNames []string
var tmpDir string

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n[compose2exe] Stopping all containers...")
		stopAll()
		cleanup()
		os.Exit(0)
	}()

	// Preflight checks
	if err := preflight(); err != nil {
		fmt.Fprintln(os.Stderr, "\n❌", err)
		fmt.Fprintln(os.Stderr, "\n[compose2exe] Aborting.")
		os.Exit(1)
	}

	// Extract embedded file volumes to temp dir
	if err := extractVolumes(); err != nil {
		fmt.Fprintln(os.Stderr, "[compose2exe] Error extracting volumes:", err)
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "[compose2exe] Error:", err)
		stopAll()
		cleanup()
		os.Exit(1)
	}

	fmt.Println("[compose2exe] All services running. Press Ctrl+C to stop.")
	select {}
}

func preflight() error {
	fmt.Println("[compose2exe] Checking environment...")

	// Check Docker is installed
	out, err := exec.Command("docker", "--version").Output()
	if err != nil {
		return fmt.Errorf("Docker is not installed or not found in PATH.\n  → Please install Docker Desktop from https://www.docker.com/products/docker-desktop/")
	}
	fmt.Println(" ✅ Docker found:", strings.TrimSpace(string(out)))

	// Check Docker daemon is running
	ping := exec.Command("docker", "info")
	ping.Stdout = io.Discard
	ping.Stderr = io.Discard
	if ping.Run() != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("Docker Desktop is not running.\n  → Please open Docker Desktop and wait for it to start, then try again.")
		}
		return fmt.Errorf("Docker daemon is not running.\n  → Please run: sudo service docker start")
	}
	fmt.Println(" ✅ Docker daemon is running")

	// Check required ports
	ports := []string{ {{range .AllPorts}}"{{.}}", {{end}} }
	for _, port := range ports {
		p := port
		if strings.Contains(p, ":") {
			p = strings.Split(p, ":")[0]
		}
		ln, err := net.Listen("tcp", ":"+p)
		if err != nil {
			return fmt.Errorf("Port %s is already in use.\n  → Please stop the process using port %s and try again.", p, p)
		}
		ln.Close()
		fmt.Printf(" ✅ Port %s is available\n", p)
	}

	fmt.Println("[compose2exe] Environment OK.")
	return nil
}

func extractVolumes() error {
	{{- if .HasFileVolumes}}
	var err error
	tmpDir, err = os.MkdirTemp("", "compose2exe-*")
	if err != nil {
		return err
	}
	{{- range .FileVolumes}}
	path_{{.SafeName}} := filepath.Join(tmpDir, "{{.SafeName}}")
	if err := os.WriteFile(path_{{.SafeName}}, fileVolume_{{.SafeName}}, 0644); err != nil {
		return fmt.Errorf("cannot extract {{.OriginalName}}: %w", err)
	}
	fmt.Println("[compose2exe] Extracted volume file: {{.OriginalName}}")
	_ = path_{{.SafeName}}
	{{- end}}
	{{- end}}
	return nil
}

func cleanup() {
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}

func run() error {
	// Create networks
{{- range $name, $net := .Networks}}
	fmt.Println("[compose2exe] Creating network: {{$name}}")
	exec.Command("docker", "network", "create", "{{$name}}").Run()
{{- end}}

	// Start services in dependency order
{{- range .Order}}{{$svcName := .}}{{$sn := sanitize .}}{{with index $.Compose.Services .}}{{if .Image}}
	if err := start_{{$sn}}(); err != nil {
		return fmt.Errorf("service {{$svcName}}: %w", err)
	}
{{- end}}{{end}}{{end}}
	return nil
}

{{range .Order}}{{$svcName := .}}{{$sn := sanitize .}}{{with index $.Compose.Services .}}{{if or .Image .Build}}
func start_{{$sn}}() error {
	image := "{{.Image}}"
	fmt.Printf("[compose2exe] Starting: {{$svcName}} (%s)\n", image)

	check := exec.Command("docker", "inspect", "--type=image", image)
	check.Stdout = io.Discard
	check.Stderr = io.Discard
	if check.Run() != nil {
{{- if $.Embed}}
		fmt.Println("[compose2exe] Loading embedded image for {{$svcName}}...")
		load := exec.Command("docker", "load")
		load.Stdin = bytes.NewReader(embedded_{{$sn}})
		load.Stdout = os.Stdout
		load.Stderr = os.Stderr
		if err := load.Run(); err != nil {
			return fmt.Errorf("docker load: %w", err)
		}
{{- else}}
		fmt.Printf("[compose2exe] Pulling: %s\n", image)
		pull := exec.Command("docker", "pull", image)
		pull.Stdout = os.Stdout
		pull.Stderr = os.Stderr
		if err := pull.Run(); err != nil {
			return fmt.Errorf("docker pull: %w", err)
		}
{{- end}}
	}

	// Check if container already running
	checkContainer := exec.Command("docker", "inspect", "--type=container", "{{$svcName}}")
	checkContainer.Stdout = io.Discard
	checkContainer.Stderr = io.Discard
	if checkContainer.Run() == nil {
		fmt.Println("[compose2exe] Container {{$svcName}} already running, skipping.")
		containerNames = append(containerNames, "{{$svcName}}")
		return nil
	}

	// Remove stopped container with same name if exists
	exec.Command("docker", "rm", "-f", "{{$svcName}}").Run()

	args := []string{"run", "-d", "--name", "{{$svcName}}"}
{{- range .Ports}}
	args = append(args, "-p", "{{.}}")
{{- end}}
{{- range $k, $v := .EnvMap}}
	args = append(args, "-e", "{{$k}}={{$v}}")
{{- end}}
{{- range .Volumes}}{{$vol := .}}
	{{- range $.FileVolumes}}{{if eq .OriginalVolume $vol}}
	args = append(args, "-v", filepath.Join(tmpDir, "{{.SafeName}}")+":{{.MountPath}}")
	{{- else}}{{end}}{{end}}
	{{- range $.FileVolumes}}{{if eq .OriginalVolume $vol}}{{else}}
	{{- end}}{{end}}
{{- end}}
{{- range .PlainVolumes}}
	args = append(args, "-v", "{{.}}")
{{- end}}
{{- range .NetworkList}}
	args = append(args, "--network", "{{.}}")
{{- end}}
{{- if .Hostname}}
	args = append(args, "--hostname", "{{.Hostname}}")
{{- end}}
{{- if .Restart}}
	args = append(args, "--restart", "{{.Restart}}")
{{- end}}
	args = append(args, image)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	containerNames = append(containerNames, "{{$svcName}}")
	return nil
}
{{end}}{{end}}{{end}}

func stopAll() {
	for i := len(containerNames) - 1; i >= 0; i-- {
		fmt.Printf("[compose2exe] Stopping %s...\n", containerNames[i])
		exec.Command("docker", "stop", containerNames[i]).Run()
		exec.Command("docker", "rm", "-f", containerNames[i]).Run()
	}
}
`
