package main

import (
	"fmt"
	"os"
	"path"

	"github.com/compose2exe/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	var composePath, name, output, module string
	var embed bool
	var targets []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--compose", "-c":
			i++; composePath = args[i]
		case "--name", "-n":
			i++; name = args[i]
		case "--embed":
			embed = true
		case "--output", "-o":
			i++; output = args[i]
		case "--target", "-t":
			i++; targets = append(targets, args[i])
		case "--module":
			i++; module = args[i]
		case "--help", "-h":
			printHelp(); os.Exit(0)
		}
	}

	if composePath == "" || name == "" {
		fmt.Fprintln(os.Stderr, "Error: --compose and --name are required")
		printHelp()
		os.Exit(1)
	}

	if output == "" {
		cwd, _ := os.Getwd()
		output = path.Join(cwd, "dist")
	}
	if module == "" {
		module = fmt.Sprintf("github.com/user/%s", name)
	}
	if len(targets) == 0 {
		targets = []string{"linux/amd64", "windows/amd64", "darwin/amd64", "darwin/arm64"}
	}

	generator := cmd.Generator{
		ComposePath: composePath,
		Name:        name,
		Embed:       embed,
		Output:      output,
		Targets:     targets,
		Module:      module,
	}

	if err := generator.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`compose2exe - Convert a Docker Compose file into a single executable

Usage:
  compose2exe --compose docker-compose.yml --name myapp [options]

Options:
  --compose, -c   Path to docker-compose.yml (required)
  --name, -n      Name of the output executable (required)
  --embed         Embed all Docker images into the binary
  --output, -o    Output directory (default: ./dist)
  --target, -t    Target platform, e.g. linux/amd64 (can repeat)
  --module        Go module name
  --help, -h      Show this help`)
}
