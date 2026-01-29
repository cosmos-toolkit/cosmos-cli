package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Template struct {
	Name     string            `yaml:"name"`
	Version  string            `yaml:"version"`
	Types    []string          `yaml:"types"`
	Defaults map[string]string `yaml:"defaults"`
	Prompts  []Prompt          `yaml:"prompts"`
	Features []string          `yaml:"features"`
	Files    FileConfig        `yaml:"files"`
}

type Prompt struct {
	Key         string `yaml:"key"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type FileConfig struct {
	Engine string `yaml:"engine"`
}

func LoadFromFS(fsys fs.FS) (*Template, error) {
	data, err := fs.ReadFile(fsys, "template.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read template.yaml: %w", err)
	}

	return LoadFromBytes(data)
}

func LoadFromPath(path string) (*Template, error) {
	data, err := os.ReadFile(filepath.Join(path, "template.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read template.yaml: %w", err)
	}

	return LoadFromBytes(data)
}

func LoadFromBytes(data []byte) (*Template, error) {
	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template.yaml: %w", err)
	}

	if err := validateTemplate(&template); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	return &template, nil
}

func validateTemplate(t *Template) error {
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if t.Files.Engine == "" {
		return fmt.Errorf("files.engine is required")
	}

	return nil
}

func (t *Template) SupportsType(typeName string) bool {
	if len(t.Types) == 0 {
		return true // No types specified means supports all
	}

	for _, t := range t.Types {
		if t == typeName {
			return true
		}
	}
	return false
}
