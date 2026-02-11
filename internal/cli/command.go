package cli

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cosmos-toolkit/cli/internal/catalog"
	"github.com/cosmos-toolkit/cli/internal/github"
	"github.com/cosmos-toolkit/cli/internal/loader"
	"github.com/cosmos-toolkit/cli/internal/pkginstall"
	"github.com/cosmos-toolkit/cli/internal/renderer"
	"github.com/cosmos-toolkit/cli/internal/resolver"
	"github.com/cosmos-toolkit/cli/internal/rules"
	"github.com/cosmos-toolkit/cli/internal/writer"
	"github.com/olekukonko/tablewriter"
)

const version = "0.1.4"

const banner = `
╭─────────────────────────────────────────────────────────╮
│                                                         │
│   ██████╗ ██████╗ ███████╗███╗   ███╗ ██████╗ ███████╗  │
│  ██╔════╝██╔═══██╗██╔════╝████╗ ████║██╔═══██╗██╔════╝  │
│  ██║     ██║   ██║███████╗██╔████╔██║██║   ██║███████╗  │
│  ██║     ██║   ██║╚════██║██║╚██╔╝██║██║   ██║╚════██║  │
│  ╚██████╗╚██████╔╝███████║██║ ╚═╝ ██║╚██████╔╝███████║  │
│   ╚═════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝  │
│                                                         │
│            From chaos to scalable structure.            │
│                                                         │
╰─────────────────────────────────────────────────────────╯
`

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

	switch args[0] {
	case "list":
		return executeList(args[1:])
	case "init":
		return executeInitCommand(args[1:])
	case "pkg":
		return executePkg(args[1:])
	case "update":
		return executeUpdate(args[1:])
	case "cache":
		return executeCache(args[1:])
	default:
		return fmt.Errorf("unknown command: %s\n\nRun 'cosmos --help' for usage", args[0])
	}
}

func executeInitCommand(args []string) error {
	// cosmos init --help or cosmos init -h
	if len(args) >= 1 && (args[0] == "--help" || args[0] == "-h") {
		printInitUsage(os.Stdout)
		return nil
	}

	// cosmos init --list or cosmos init -l
	if len(args) >= 1 && (args[0] == "--list" || args[0] == "-l") {
		printTemplateList(os.Stdout)
		return nil
	}

	// cosmos init (no args) or cosmos init --interactive/-i -> interactive mode
	initArgs := args
	if len(initArgs) == 0 || (len(initArgs) == 1 && (initArgs[0] == "--interactive" || initArgs[0] == "-i")) {
		return runInteractiveInit()
	}

	config, err := parseInitCommand(initArgs)
	if err != nil {
		if err == flag.ErrHelp {
			printInitUsage(os.Stdout)
			return nil
		}
		return err
	}

	return executeInit(config)
}

func executeList(args []string) error {
	// cosmos list or cosmos list --help
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printListUsage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "templates":
		return runListTemplates(os.Stdout)
	case "pkgs", "packages":
		return runListPackages(os.Stdout)
	default:
		return fmt.Errorf("unknown subcommand: %s\n\nRun 'cosmos list --help' for usage", args[0])
	}
}

func runListTemplates(w io.Writer) error {
	printBanner(w)
	fmt.Fprintf(w, "%s\n\n", title("Available templates"))
	fmt.Fprintf(w, "%s\n\n", dimmed("github.com/cosmos-toolkit/templates"))

	templates, err := github.ListTemplatesWithInfo()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}
	if len(templates) == 0 {
		fmt.Fprintf(w, "  %s\n", dimmed("(no templates yet)"))
		fmt.Fprintln(w)
		return nil
	}

	data := [][]string{{"NAME", "DESCRIPTION", "LINK"}}
	for _, t := range templates {
		data = append(data, []string{t.Name, t.Description, t.Link})
	}

	table := tablewriter.NewWriter(w)
	table.Header(data[0])
	table.Bulk(data[1:])
	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}
	fmt.Fprintln(w)
	return nil
}

func runListPackages(w io.Writer) error {
	printBanner(w)
	fmt.Fprintf(w, "%s\n\n", title("Available packages"))
	fmt.Fprintf(w, "%s\n\n", dimmed("github.com/cosmos-toolkit/packages"))

	pkgs, err := github.ListPackagesWithInfo()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}
	if len(pkgs) == 0 {
		fmt.Fprintf(w, "  %s\n", dimmed("(no packages yet)"))
		fmt.Fprintln(w)
		return nil
	}

	data := [][]string{{"NAME", "DESCRIPTION", "LINK"}}
	for _, p := range pkgs {
		data = append(data, []string{p.Name, p.Description, p.Link})
	}

	table := tablewriter.NewWriter(w)
	table.Header(data[0])
	table.Bulk(data[1:])
	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}
	fmt.Fprintln(w)
	return nil
}

func executeUpdate(args []string) error {
	if len(args) >= 1 && (args[0] == "--help" || args[0] == "-h") {
		printUpdateUsage(os.Stdout)
		return nil
	}
	okT, err := resolver.PullTemplatesRepo()
	if err != nil {
		return err
	}
	if okT {
		fmt.Printf("%s Templates cache updated\n", green+"✓"+reset)
	}
	okP, err := resolver.PullPackagesRepo()
	if err != nil {
		return err
	}
	if okP {
		fmt.Printf("%s Packages cache updated\n", green+"✓"+reset)
	}
	if !okT && !okP {
		fmt.Println(dimmed("No cache found. Use 'cosmos init' or 'cosmos pkg' to create it."))
	}
	return nil
}

func executeCache(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printCacheUsage(os.Stdout)
		return nil
	}
	if args[0] != "refresh" {
		return fmt.Errorf("unknown subcommand: %s\n\nRun 'cosmos cache --help' for usage", args[0])
	}
	return executeUpdate(args[1:])
}

func executePkg(args []string) error {
	if len(args) >= 1 && (args[0] == "--help" || args[0] == "-h") {
		printPkgUsage(os.Stdout)
		return nil
	}

	force, positionals := parsePkgArgs(args)

	// cosmos pkg (no positionals) or only -i/--interactive -> interactive mode
	if len(positionals) == 0 {
		return runInteractivePkg(force)
	}

	name := positionals[0]
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	pkgDir := filepath.Join(cwd, "pkg", name)
	if writer.DirectoryExists(pkgDir) && force {
		fmt.Printf("%s Overwriting existing %s\n", dimmed("→"), dimmed("pkg/"+name))
	}

	opts := pkginstall.InstallOpts{Force: force}
	if err := pkginstall.Install(name, cwd, opts); err != nil {
		return err
	}

	fmt.Printf("%s Package %s installed in %s/pkg/%s\n", green+"✓"+reset, accent(name), dimmed(cwd), accent(name))
	return nil
}

// parsePkgArgs extracts --force/-f and returns (force, positionals).
// Positionals are args that are not --help, -h, -i, --interactive, --force, -f.
func parsePkgArgs(args []string) (force bool, positionals []string) {
	for _, a := range args {
		switch a {
		case "--force", "-f":
			force = true
		case "--help", "-h", "--interactive", "-i":
			// skip
		default:
			positionals = append(positionals, a)
		}
	}
	return force, positionals
}

func printPkgUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

  %s pkg                    Interactive: list packages, select one or more to install
  %s pkg %s       Install a package into the current project
  %s pkg %s       List available packages

  Run from the root of your Go project (where go.mod is).
  The package and its copy_deps are copied to pkg/<name> and imports
  are rewritten to your module path. Dependencies are added with go get.

  If pkg/<name> already exists, the command fails unless %s is used.
  With %s, existing content is overwritten (useful for automation).

%s
  %s, %s
      Overwrite existing pkg/<name> if it exists (fails by default)

%s
  %s %s pkg
  %s %s pkg %s
  %s %s pkg %s
  %s %s pkg %s
  %s %s pkg %s
  %s %s pkg %s %s

`,
		title("Install a reusable package into the current project."),
		cmd("cosmos"),
		cmd("cosmos"), accent("<name>"),
		cmd("cosmos"), accent("list pkgs"),
		flagStyle("--force"),
		flagStyle("--force"),
		section("FLAGS:"),
		flagStyle("--force"), flagStyle("-f"),
		section("EXAMPLES:"),
		dimmed("#"), cmd("cosmos"),
		dimmed("#"), cmd("cosmos"), flagStyle("-i"),
		dimmed("#"), cmd("cosmos"), accent("logger"),
		dimmed("#"), cmd("cosmos"), accent("config"),
		dimmed("#"), cmd("cosmos"), accent("validator"),
		dimmed("#"), cmd("cosmos"), accent("logger"), flagStyle("--force"),
	)
}

func printListUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

  %s list %s      List available templates (from github.com/cosmos-toolkit/templates)
  %s list %s      List available packages (from github.com/cosmos-toolkit/packages)

`,
		title("List templates and packages."),
		cmd("cosmos"), accent("templates"),
		cmd("cosmos"), accent("pkgs"),
	)
}

func printBanner(w io.Writer) {
	lines := strings.Split(strings.TrimSpace(banner), "\n")
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w)
}

func printUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

  %s init              Start a new project (interactive)
  %s pkg               Install packages (interactive: list, select one or more)
  %s pkg %s       Install a package (logger, config, ...) into current project
  %s update            Refresh templates and packages cache (git pull)
  %s cache %s    Same as %s update
  %s list %s     List available templates
  %s list %s     List available packages

  %s %s, %s    Show this help
  %s %s, %s    Show version

`,
		section("USAGE:"),
		cmd("cosmos"),
		cmd("cosmos"),
		cmd("cosmos"), accent("<name>"),
		cmd("cosmos"),
		cmd("cosmos"), accent("refresh"),
		cmd("cosmos"),
		cmd("cosmos"), accent("templates"),
		cmd("cosmos"), accent("pkgs"),
		cmd("cosmos"), flagStyle("--help"), flagStyle("-h"),
		cmd("cosmos"), flagStyle("--version"), flagStyle("-v"),
	)
}

func printUpdateUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

  %s update    Run git pull in the local templates and packages caches
               (~/.cache/cosmos/templates/_repo and ~/.cache/cosmos/packages/_repo)
               so the next init or pkg uses the latest versions.

  If a cache does not exist yet, nothing is done for that cache.
  Use %s init or %s pkg to create the cache on first use.

`,
		title("Refresh templates and packages cache."),
		cmd("cosmos"),
		cmd("cosmos"),
		cmd("cosmos"),
	)
}

func printCacheUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

  %s cache %s    Same as %s update: refresh templates and packages cache (git pull).

`,
		title("Cache operations."),
		cmd("cosmos"), accent("refresh"),
		cmd("cosmos"),
	)
}

func printInitUsage(w io.Writer) {
	printBanner(w)
	fmt.Fprintf(w, `%s

%s
  %s init                        Interactive setup (project name, template, module)
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
      External template name. Fetched from github.com/cosmos-toolkit/templates/<name>
      Cached under ~/.cache/cosmos/templates/
  %s
      Overwrite existing project directory if it exists
  %s, %s
      List available built-in and external templates

%s
  %s List available templates
  %s init %s

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
		cmd("cosmos"), cmd("cosmos"),
		cmd("cosmos"), flagStyle("--template <name>"),
		cmd("cosmos"), flagStyle("--module <module-path>"),
		cmd("cosmos"), flagStyle("--template <template-name>"), flagStyle("--module <module-path>"),
		section("ARGUMENTS:"), flagStyle("--template"),
		section("FLAGS:"),
		flagStyle("--module"), flagStyle("--template"), flagStyle("--force"),
		flagStyle("--list"), flagStyle("-l"),
		section("EXAMPLES:"),
		dimmed("#"), cmd("cosmos"), flagStyle("--list"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"), flagStyle("--template"),
		dimmed("#"), cmd("cosmos"), flagStyle("--module"), flagStyle("--force"),
	)
}

var builtInDescriptions = map[string]string{
	"api":    "HTTP service with handlers and server",
	"worker": "Background processing / async jobs",
	"cli":    "Command-line tool with subcommands",
}

func printTemplateList(w io.Writer) {
	printBanner(w)
	cat := catalog.New()
	builtIn := cat.ListTemplates()

	fmt.Fprintf(w, "%s\n\n", title("Available templates"))
	fmt.Fprintf(w, "%s\n", section("Built-in templates (use directly):"))
	fmt.Fprintf(w, "\n")
	for _, t := range builtIn {
		desc := builtInDescriptions[t.Type]
		if desc == "" {
			desc = "Go project template"
		}
		fmt.Fprintf(w, "  %s  %s\n", accent(t.Type), dimmed(desc))
		if len(t.Features) > 0 {
			fmt.Fprintf(w, "      features: %s\n", dimmed(strings.Join(t.Features, ", ")))
		}
		fmt.Fprintf(w, "      %s init %s <name> %s\n\n", cmd("cosmos"), t.Type, flagStyle("--module <path>"))
	}

	external, err := github.ListTemplatesWithInfo()
	if err != nil {
		fmt.Fprintf(w, "  %s\n\n", dimmed("(could not list external templates)"))
	} else {
		fmt.Fprintf(w, "%s\n", section("External templates (from GitHub):"))
		fmt.Fprintf(w, "\n")
		for _, t := range external {
			fmt.Fprintf(w, "  %s  %s\n", accent(t.Name), dimmed(t.Description))
			fmt.Fprintf(w, "      %s init <name> %s %s\n\n", cmd("cosmos"), flagStyle("--template "+t.Name), flagStyle("--module <path>"))
		}
	}
	fmt.Fprintf(w, "  Use %s to fetch templates from:\n  %s\n\n", flagStyle("--template <name>"), dimmed("github.com/cosmos-toolkit/templates/<name>"))
	fmt.Fprintf(w, "%s\n", dimmed("Run 'cosmos init --help' for more details."))
}

func runInteractiveInit() error {
	printBanner(os.Stdout)
	fmt.Println(title("Let's create a new Go project"))
	fmt.Println()

	var projectName string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Project name:",
			Help:    "This will be the directory name for your project",
		},
		&projectName,
		survey.WithValidator(survey.Required),
		survey.WithValidator(func(ans interface{}) error {
			s := ans.(string)
			return rules.ValidateProjectName(s)
		}),
	); err != nil {
		return err
	}

	// Build template options: built-in + external (from GitHub, descriptions from manifest)
	externalTemplates, err := github.ListTemplatesWithInfo()
	if err != nil {
		externalTemplates = nil // proceed with built-in only
	}
	templateOpts := make([]string, 0, 3+len(externalTemplates))
	for _, t := range []struct {
		name string
		desc string
	}{
		{"api", builtInDescriptions["api"]},
		{"worker", builtInDescriptions["worker"]},
		{"cli", builtInDescriptions["cli"]},
	} {
		templateOpts = append(templateOpts, fmt.Sprintf("%s - %s (built-in)", t.name, t.desc))
	}
	for _, t := range externalTemplates {
		templateOpts = append(templateOpts, fmt.Sprintf("%s - %s (external)", t.Name, t.Description))
	}

	var selectedTemplate string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Choose a template:",
			Options: templateOpts,
		},
		&selectedTemplate,
	); err != nil {
		return err
	}

	// Parse selection: "name - description (built-in)" or "name - description (external)"
	parts := strings.SplitN(selectedTemplate, " - ", 2)
	templateID := strings.TrimSpace(parts[0])
	isExternal := strings.Contains(selectedTemplate, "(external)")

	var modulePath string
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "user"
	}
	defaultModule := fmt.Sprintf("github.com/%s/%s", user, projectName)
	if err := survey.AskOne(
		&survey.Input{
			Message: "Module path:",
			Help:    "Go module path (e.g. github.com/user/repo)",
			Default: defaultModule,
		},
		&modulePath,
		survey.WithValidator(survey.Required),
		survey.WithValidator(func(ans interface{}) error {
			s := ans.(string)
			return rules.ValidateModulePath(s)
		}),
	); err != nil {
		return err
	}

	config := &Config{
		ProjectName: projectName,
		Module:      modulePath,
	}
	if isExternal {
		config.Template = templateID
	} else {
		config.Type = templateID
	}

	// Confirm overwrite if directory exists
	if writer.DirectoryExists(projectName) {
		var overwrite bool
		if err := survey.AskOne(
			&survey.Confirm{
				Message: fmt.Sprintf("Directory %q already exists. Overwrite?", projectName),
				Default: false,
			},
			&overwrite,
		); err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("cancelled: directory %s already exists", projectName)
		}
		config.Force = true
	}

	return executeInit(config)
}

func runInteractivePkg(force bool) error {
	printBanner(os.Stdout)
	fmt.Println(title("Install packages into the current project"))
	fmt.Println()

	pkgs, err := github.ListPackagesWithInfo()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages available")
	}

	options := make([]string, len(pkgs))
	for i, p := range pkgs {
		desc := strings.TrimSpace(p.Description)
		if desc == "" || desc == "-" {
			options[i] = p.Name
		} else {
			options[i] = fmt.Sprintf("%s - %s", p.Name, p.Description)
		}
	}

	var selected []string
	if err := survey.AskOne(
		&survey.MultiSelect{
			Message: "Choose one or more packages to install (space to select, enter to confirm):",
			Options: options,
		},
		&selected,
		survey.WithValidator(survey.MinItems(1)),
	); err != nil {
		return err
	}

	// Map selected option strings back to package names (first part before " - ")
	names := make([]string, 0, len(selected))
	for _, opt := range selected {
		parts := strings.SplitN(opt, " - ", 2)
		names = append(names, strings.TrimSpace(parts[0]))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check which selected packages already exist
	var existing []string
	for _, name := range names {
		if writer.DirectoryExists(filepath.Join(cwd, "pkg", name)) {
			existing = append(existing, name)
		}
	}
	if len(existing) > 0 && !force {
		var overwrite bool
		if err := survey.AskOne(
			&survey.Confirm{
				Message: fmt.Sprintf("The following already exist in pkg/: %s. Overwrite?", strings.Join(existing, ", ")),
				Default: false,
			},
			&overwrite,
		); err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("skipped: use --force to overwrite existing packages")
		}
		force = true
	}

	opts := pkginstall.InstallOpts{Force: force}
	for _, name := range names {
		pkgDir := filepath.Join(cwd, "pkg", name)
		if writer.DirectoryExists(pkgDir) && force {
			fmt.Printf("%s Overwriting existing %s\n", dimmed("→"), dimmed("pkg/"+name))
		}
		if err := pkginstall.Install(name, cwd, opts); err != nil {
			return fmt.Errorf("failed to install %s: %w", name, err)
		}
		fmt.Printf("%s Package %s installed in %s/pkg/%s\n", green+"✓"+reset, accent(name), dimmed(cwd), accent(name))
	}

	return nil
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

	modulePlaceholder := template.Files.ModulePlaceholder
	if modulePlaceholder == "" && config.Template != "" {
		modulePlaceholder = "github.com/your-org/your-app"
	}

	ctx := renderer.Context{
		ProjectName:       config.ProjectName,
		Module:            config.Module,
		GoVersion:         goVersion,
		ModulePlaceholder: modulePlaceholder,
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
