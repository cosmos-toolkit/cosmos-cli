package cli

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cosmos-toolkit/cosmos-cli/internal/catalog"
	"github.com/cosmos-toolkit/cosmos-cli/internal/loader"
	"github.com/cosmos-toolkit/cosmos-cli/internal/renderer"
	"github.com/cosmos-toolkit/cosmos-cli/internal/resolver"
	"github.com/cosmos-toolkit/cosmos-cli/internal/rules"
	"github.com/cosmos-toolkit/cosmos-cli/internal/writer"
)

type Config struct {
	Type        string
	ProjectName string
	Module      string
	Template    string
	Force       bool
}

func Execute() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: cosmos init [type] <name> [flags] or cosmos init <name> --template <name>")
	}

	if os.Args[1] != "init" {
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}

	config, err := parseInitCommand(os.Args[2:])
	if err != nil {
		return err
	}

	return executeInit(config)
}

func executeInit(config *Config) error {
	// Validate inputs
	if err := rules.ValidateModulePath(config.Module); err != nil {
		return err
	}

	if err := rules.ValidateProjectName(config.ProjectName); err != nil {
		return err
	}

	// Determine output directory
	outputDir := config.ProjectName
	if writer.DirectoryExists(outputDir) && !config.Force {
		return fmt.Errorf("directory %s already exists. Use --force to overwrite", outputDir)
	}

	if config.Force && writer.DirectoryExists(outputDir) {
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Resolve template
	var templateFS fs.FS
	var template *loader.Template
	var err error

	if config.Template != "" {
		// External template
		if err := rules.ValidateTemplateName(config.Template); err != nil {
			return err
		}

		templatePath, err := resolver.Resolve(config.Template)
		if err != nil {
			return fmt.Errorf("failed to resolve template: %w", err)
		}

		template, err = loader.LoadFromPath(templatePath)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		templateFS = os.DirFS(templatePath)
	} else {
		// Embedded template
		if config.Type == "" {
			return fmt.Errorf("either specify a type (api, worker, cli) or use --template")
		}

		if err := rules.ValidateType(config.Type); err != nil {
			return err
		}

		cat := catalog.New()
		embeddedFS, ok := cat.GetEmbeddedTemplate(config.Type)
		if !ok {
			return fmt.Errorf("template type %s not found", config.Type)
		}

		template, err = loader.LoadFromFS(embeddedFS)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Validate type compatibility
		if err := rules.ValidateTypeCompatibility(template.Types, config.Type); err != nil {
			return err
		}

		templateFS = embeddedFS
	}

	// Prepare render context
	goVersion := template.Defaults["goVersion"]
	if goVersion == "" {
		goVersion = "1.23"
	}

	ctx := renderer.Context{
		ProjectName: config.ProjectName,
		Module:      config.Module,
		GoVersion:   goVersion,
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Render template
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := renderer.Render(templateFS, ctx, absOutputDir); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	fmt.Printf("✓ Project %s initialized successfully!\n", config.ProjectName)
	return nil
}

func parseInitCommand(args []string) (*Config, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	module := fs.String("module", "", "Go module path (required)")
	template := fs.String("template", "", "External template name")
	force := fs.Bool("force", false, "Overwrite existing directory")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return nil, fmt.Errorf("project name is required")
	}

	var config Config

	// Check if first arg is a type (api, worker, cli) or project name
	if len(remaining) == 1 {
		// Only project name provided - requires --template
		if *template == "" {
			return nil, fmt.Errorf("either specify a type (api, worker, cli) or use --template")
		}
		config.ProjectName = remaining[0]
	} else if len(remaining) == 2 {
		// Could be: type + name OR name + something else
		// Check if first is a valid type
		if isValidType(remaining[0]) {
			config.Type = remaining[0]
			config.ProjectName = remaining[1]
		} else {
			return nil, fmt.Errorf("invalid type: %s. Valid types: api, worker, cli", remaining[0])
		}
	} else {
		return nil, fmt.Errorf("too many arguments")
	}

	if *module == "" {
		return nil, fmt.Errorf("--module is required")
	}

	config.Module = *module
	config.Template = *template
	config.Force = *force

	return &config, nil
}

func isValidType(t string) bool {
	return t == "api" || t == "worker" || t == "cli"
}
