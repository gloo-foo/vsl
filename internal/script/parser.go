// Package script provides utilities for parsing UP script files.
package script

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gloo-foo/vsl/internal/container"
	runpkg "github.com/gloo-foo/vsl/internal/container/run"
	up "github.com/uplang/go"
)

// ParseFile parses an UP script file and returns the configuration.
func ParseFile(path string) (*runpkg.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Read and skip shebang
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		// Remove shebang line
		content = []byte(strings.Join(lines[1:], "\n"))
	}

	// Parse UP document
	parser := up.NewParser()
	doc, err := parser.ParseDocument(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse UP document: %w", err)
	}

	config := &runpkg.Config{
		Command:     []container.Command{},
		Entrypoint:  []container.Entrypoint{},
		Environment: []container.Environment{},
		Volumes:     []container.Volume{},
	}

	// Extract values from UP document
	for _, node := range doc.Nodes {
		switch node.Key {
		case "image":
			if scalar, ok := node.Value.(string); ok {
				config.Image = container.Image(scalar)
			}
		case "command":
			for _, cmd := range extractList(node.Value) {
				config.Command = append(config.Command, container.Command(cmd))
			}
		case "entrypoint":
			for _, ep := range extractList(node.Value) {
				config.Entrypoint = append(config.Entrypoint, container.Entrypoint(ep))
			}
		case "workdir", "working_dir":
			if scalar, ok := node.Value.(string); ok {
				config.WorkingDir = container.WorkingDir(scalar)
			}
		case "env", "environment":
			for _, env := range extractEnvironment(node.Value) {
				config.Environment = append(config.Environment, container.Environment(env))
			}
		case "volume", "volumes":
			for _, vol := range extractList(node.Value) {
				config.Volumes = append(config.Volumes, container.Volume(vol))
			}
		case "user":
			if scalar, ok := node.Value.(string); ok {
				config.User = container.User(scalar)
			}
		case "interactive":
			if scalar, ok := node.Value.(string); ok {
				config.Interactive = string(scalar) == "true"
			}
		case "privileged":
			if scalar, ok := node.Value.(string); ok {
				config.Privileged = string(scalar) == "true"
			}
		case "network_mode":
			if scalar, ok := node.Value.(string); ok {
				config.NetworkMode = container.NetworkMode(scalar)
			}
		}
	}

	if config.Image == "" {
		return nil, fmt.Errorf("script must specify image")
	}

	return config, nil
}

func extractList(value up.Value) []string {
	if list, ok := value.(up.List); ok {
		result := make([]string, 0, len(list))
		for _, item := range list {
			if scalar, ok := item.(string); ok {
				result = append(result, scalar)
			}
		}
		return result
	}
	return nil
}

func extractEnvironment(value up.Value) []string {
	var result []string

	switch v := value.(type) {
	case up.List:
		// Array form: ["KEY=value", "KEY2=value2"]
		for _, item := range v {
			if scalar, ok := item.(string); ok {
				result = append(result, scalar)
			}
		}
	case up.Block:
		// Map form: { KEY value, KEY2 value2 }
		for key, val := range v {
			if scalar, ok := val.(string); ok {
				result = append(result, fmt.Sprintf("%s=%s", key, scalar))
			}
		}
	}

	return result
}
