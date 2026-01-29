package cli

import (
	"flag"
	"fmt"
	"io"
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

const version = "0.1.0"

type Config struct {
	Type        string
	ProjectName string
	Module      string
	Template    string
	Force       bool
}

func Execute() error {
	args := os.Args[1:]

	// No args or --help/-h at top level
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}
	if args[0] == "--help" || args[0] == "-h" {
		printUsage(os.Stdout)
		return nil
	}
	if args[0] == "version" || args[0] == "--version" || args[0] == "-v" {
		fmt.Printf("%s %s %s\n", cmd("cosmos"), dimmed("version"), accent(version))
		return nil
	}

	if args[0] != "init" {
		return fmt.Errorf("unknown command: %s\n\nRun 'cosmos --help' for usage", args[0])
	}

	// cosmos init --help or cosmos init -h
	if len(args) >= 2 && (args[1] == "--help" || args[1] == "-h") {
		printInitUsage(os.Stdout)
		return nil
	}

	config, err := parseInitCommand(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			printInitUsage(os.Stdout)
			return nil
		}
		return err
	}

	return executeInit(config)
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `%s  Initialize Go projects from templates.

%s
  %s init [type] <name> [flags]   Create project with built-in template
  %s init <name> %s  Create project with external template
  %s version                    Show version
  %s %s                     Show this help

%s
  init    Initialize a new Go project from a template

%s
  api     HTTP service
  worker  Background processing / async jobs
  cli     Command-line tool

%s
  %s init api payments %s github.com/user/payments
  %s init worker jobs %s github.com/user/jobs
  %s init myapp %s github.com/user/myapp %s hexagonal-architecture

%s
  %s string    Go module path (required)
  %s string  External template name (from github.com/cosmos-cli/templates/<name>)
  %s            Overwrite existing directory

%s
`,
		title("Cosmos —"),
		section("USAGE:"),
		cmd("cosmos"), cmd("cosmos"), flagStyle("--template <n>"),
		cmd("cosmos"), cmd("cosmos"), flagStyle("--help"),
		section("COMMANDS:"),
		section("BUILT-IN TYPES (use with init):"),
		section("EXAMPLES:"),
		cmd("cosmos"), flagStyle("--module"),
		cmd("cosmos"), flagStyle("--module"),
		cmd("cosmos"), flagStyle("--module"), flagStyle("--template"),
		section("FLAGS (init):"),
		flagStyle("--module"), flagStyle("--template"), flagStyle("--force"),
		dimmed("Run 'cosmos init --help' for init command details."),
	)
}

func printInitUsage(w io.Writer) {
	fmt.Fprintf(w, `%s

%s
  %s init [type] <name> [flags]
  %s init <name> %s [flags]

  With built-in type (api, worker, cli):
    %s init <type> <project-name> %s

  With external template (from GitHub):
    %s init <project-name> %s %s

%s
  type          One of: api, worker, cli (required when not using %s)
  project-name  Name of the project directory to create
  template-name Name of external template (e.g. hexagonal-architecture)

%s
  %s string
      Go module path (required). Example: github.com/user/repo
  %s string
      External template name. Fetched from github.com/cosmos-cli/templates/<name>
      Cached under ~/.cache/cosmos/templates/
  %s
      Overwrite existing project directory if it exists

%s
  %s API project
  %s init api payments %s github.com/myorg/payments

  %s Worker project
  %s init worker jobs %s github.com/myorg/jobs

  %s CLI tool
  %s init cli toolbox %s github.com/myorg/toolbox

  %s External template (e.g. DDD, Hexagonal)
  %s init myapp %s github.com/myorg/myapp %s ddd-architecture

  %s Overwrite existing directory
  %s init api payments %s github.com/myorg/payments %s
`,
		title("Initialize a new Go project from a template."),
		section("USAGE:"),
		cmd("cosmos"), cmd("cosmos"), flagStyle("--template <name>"),
		cmd("cosmos"), flagStyle("--module <module-path>"),
		cmd("cosmos"), flagStyle("--template <template-name>"), flagStyle("--module <module-path>"),
		section("ARGUMENTS:"), flagStyle("--template"),
		section("FLAGS:"),
		flagStyle("--module"), flagStyle("--template"), flagStyle("--force"),
		section("EXAMPLES:"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"), flagStyle("--template"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"), flagStyle("--force"),
	)
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

	fmt.Printf("%s Project %s initialized successfully!\n", green+"✓"+reset, accent(config.ProjectName))
	return nil
}

func parseInitCommand(args []string) (*Config, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.Usage = func() { printInitUsage(os.Stdout) }
	module := fs.String("module", "", "Go module path (required)")
	template := fs.String("template", "", "External template name")
	force := fs.Bool("force", false, "Overwrite existing directory")

	if len(args) == 0 {
		return nil, fmt.Errorf("project name is required")
	}

	var config Config
	var flagArgs []string

	// Parse positionals first so flags can appear after them (e.g. cosmos init api myapp --module x).
	// Go's flag package stops at the first non-flag, so we must split positionals from flag args ourselves.
	if len(args) >= 2 && isValidType(args[0]) {
		config.Type = args[0]
		config.ProjectName = args[1]
		flagArgs = args[2:]
	} else if len(args) >= 1 {
		config.ProjectName = args[0]
		flagArgs = args[1:]
	}

	if err := fs.Parse(flagArgs); err != nil {
		return nil, err
	}

	if config.Type == "" && *template == "" {
		return nil, fmt.Errorf("either specify a type (api, worker, cli) or use --template")
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
