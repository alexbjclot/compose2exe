package cmd



const mainTemplate = `package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
{{- if .Embed}}
	"bytes"
	_ "embed"
{{- end}}
)

{{- if .Embed}}
{{- range .Order}}{{$sn := sanitize .}}{{with index $.Compose.Services .}}{{if .Image}}
//go:embed {{$sn}}.tar.gz
var embedded_{{$sn}} []byte
{{- end}}{{end}}{{end}}
{{- end}}

var containerNames []string

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n[compose2exe] Stopping all containers...")
		stopAll()
		os.Exit(0)
	}()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "[compose2exe] Error:", err)
		stopAll()
		os.Exit(1)
	}

	fmt.Println("[compose2exe] All services running. Press Ctrl+C to stop.")
	select {}
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

{{range .Order}}{{$svcName := .}}{{$sn := sanitize .}}{{with index $.Compose.Services .}}{{if .Image}}
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
{{- range .Volumes}}
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
