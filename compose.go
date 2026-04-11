package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ComposeFile struct {
	Services map[string]*Service
	Networks map[string]*Network
	Volumes  map[string]bool
}

type Service struct {
	Image       string
	Ports       []string
	EnvMap      map[string]string
	Volumes     []string
	NetworkList []string
	DependsList []string
	Hostname    string
	Restart     string
}

type Network struct {
	External bool
}

// ParseCompose parses a docker-compose.yml using a simple line-by-line approach
func ParseCompose(filePath string) (*ComposeFile, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open compose file: %w", err)
	}
	defer f.Close()

	compose := &ComposeFile{
		Services: make(map[string]*Service),
		Networks: make(map[string]*Network),
		Volumes:  make(map[string]bool),
	}

	var (
		inServices  bool
		inNetworks  bool
		inVolumes   bool
		currentSvc  string
		currentSvcObj *Service
		inEnv       bool
		inPorts     bool
		inVolsList  bool
		inNets      bool
		inDepends   bool
		inSvcNets   bool
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		// Top-level sections
		if indent == 0 {
			inServices = strings.HasPrefix(trimmed, "services:")
			inNetworks = strings.HasPrefix(trimmed, "networks:")
			inVolumes = strings.HasPrefix(trimmed, "volumes:")
			currentSvc = ""
			currentSvcObj = nil
			inEnv = false; inPorts = false; inVolsList = false
			inNets = false; inDepends = false; inSvcNets = false
			continue
		}

		// Service names (indent 2)
		if inServices && indent == 2 && strings.HasSuffix(trimmed, ":") {
			currentSvc = strings.TrimSuffix(trimmed, ":")
			currentSvcObj = &Service{EnvMap: make(map[string]string)}
			compose.Services[currentSvc] = currentSvcObj
			inEnv = false; inPorts = false; inVolsList = false
			inNets = false; inDepends = false; inSvcNets = false
			continue
		}

		// Network names (indent 2)
		if inNetworks && indent == 2 {
			netName := strings.TrimSuffix(trimmed, ":")
			if _, exists := compose.Networks[netName]; !exists {
				compose.Networks[netName] = &Network{}
			}
			continue
		}

		// Network external flag (indent 4)
		if inNetworks && indent == 4 && strings.HasPrefix(trimmed, "external:") {
			// find which network we're in — skip for simplicity, treat all as internal
			continue
		}

		// Volume names (indent 2)
		if inVolumes && indent == 2 {
			volName := strings.TrimSuffix(trimmed, ":")
			compose.Volumes[volName] = true
			continue
		}

		// Service properties (indent 4)
		if inServices && currentSvcObj != nil && indent == 4 {
			inEnv = false; inPorts = false; inVolsList = false
			inNets = false; inDepends = false; inSvcNets = false

			if strings.HasPrefix(trimmed, "image:") {
				currentSvcObj.Image = strings.TrimSpace(strings.TrimPrefix(trimmed, "image:"))
				currentSvcObj.Image = strings.Trim(currentSvcObj.Image, "\"'")
			} else if strings.HasPrefix(trimmed, "hostname:") {
				currentSvcObj.Hostname = strings.TrimSpace(strings.TrimPrefix(trimmed, "hostname:"))
			} else if strings.HasPrefix(trimmed, "restart:") {
				currentSvcObj.Restart = strings.TrimSpace(strings.TrimPrefix(trimmed, "restart:"))
			} else if trimmed == "environment:" {
				inEnv = true
			} else if trimmed == "ports:" {
				inPorts = true
			} else if trimmed == "volumes:" {
				inVolsList = true
			} else if trimmed == "networks:" {
				inSvcNets = true
			} else if trimmed == "depends_on:" {
				inDepends = true
			} else if strings.HasPrefix(trimmed, "networks:") {
				// inline networks
			}
			continue
		}

		// Service sub-lists (indent 6)
		if inServices && currentSvcObj != nil && indent == 6 {
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.Trim(item, "\"'")

			if inEnv {
				if strings.Contains(item, "=") {
					parts := strings.SplitN(item, "=", 2)
					currentSvcObj.EnvMap[parts[0]] = parts[1]
				} else if strings.Contains(item, ":") {
					parts := strings.SplitN(item, ":", 2)
					currentSvcObj.EnvMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			} else if inPorts {
				currentSvcObj.Ports = append(currentSvcObj.Ports, item)
			} else if inVolsList {
				currentSvcObj.Volumes = append(currentSvcObj.Volumes, item)
			} else if inSvcNets {
				netName := strings.TrimSuffix(item, ":")
				currentSvcObj.NetworkList = append(currentSvcObj.NetworkList, netName)
			} else if inDepends {
				currentSvcObj.DependsList = append(currentSvcObj.DependsList, item)
			}
			_ = inNets
			continue
		}
	}

	return compose, scanner.Err()
}

// GetOrderedServices returns services sorted by dependency order
func GetOrderedServices(compose *ComposeFile) []string {
	visited := make(map[string]bool)
	var order []string

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		svc := compose.Services[name]
		if svc != nil {
			for _, dep := range svc.DependsList {
				visit(dep)
			}
		}
		order = append(order, name)
	}

	for name := range compose.Services {
		visit(name)
	}
	return order
}
