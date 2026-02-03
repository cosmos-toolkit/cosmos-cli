# 🌌 Cosmos

**From chaos to scalable structure.**

Cosmos is a CLI that helps you bootstrap Go projects from templates and add reusable packages. It focuses on clarity, structure, and long-term maintainability.

## Why Cosmos exists

New projects often start with unclear boundaries, inconsistent structure, and shortcuts that become permanent. Cosmos turns that initial chaos into **intentional structure**: an explicit, predictable starting point so systems can grow without friction.

## What Cosmos does

- **Initializes projects** from built-in templates (api, worker, cli) or external templates from `github.com/cosmos-toolkit/templates`.
- **Installs packages** from `github.com/cosmos-toolkit/packages` into your project’s `pkg/` (with import rewriting and dependency resolution).
- **Validates** project names, module paths, and template/type choices before generating anything.

Cosmos is a **system initializer**, not a framework: explicit by default, deterministic, and focused on long-term maintenance.

## Core concepts

### Built-in templates

Three template types are embedded in the CLI:

- **api** — HTTP service with handlers and server
- **worker** — background processing and async jobs
- **cli** — command-line tool with subcommands

### External templates

Additional templates live in `github.com/cosmos-toolkit/templates`. Each subdirectory is one template (e.g. `api-hexagonal`). They are listed via the GitHub API and in the interactive init menu. Descriptions come from a root `manifest.yaml` (key `templates.<name>.description`). Templates are fetched with **git sparse checkout** and cached under `~/.cache/cosmos/templates/_repo`.

### Templates as contracts

Each template has a `template.yaml` that declares:

- **name**, **version**, **types** (e.g. `["api"]`)
- **defaults** (e.g. `goVersion: "1.23"`)
- **prompts** (e.g. module, projectName)
- **features** (optional list)
- **files.engine** (e.g. `gotmpl`) and optional **files.modulePlaceholder** for import rewriting

This keeps generation predictable and maintainable.

## Commands overview

| Command                                     | Description                                                             |
| ------------------------------------------- | ----------------------------------------------------------------------- |
| `cosmos` / `cosmos --help`                  | Show usage                                                              |
| `cosmos version` / `cosmos -v`              | Show version                                                            |
| `cosmos init`                               | Interactive: project name, template (built-in or external), module path |
| `cosmos init --list` / `-l`                 | List built-in and external templates                                    |
| `cosmos list templates`                     | List external templates (from GitHub)                                   |
| `cosmos list pkgs` / `cosmos list packages` | List available packages                                                 |
| `cosmos update`                             | Refresh templates and packages caches (git pull)                        |
| `cosmos cache refresh`                      | Same as `cosmos update`                                                 |
| `cosmos pkg`                                | Interactive: select one or more packages to install                     |
| `cosmos pkg <name>`                         | Install package into current project                                    |

## Usage

### Creating a project (interactive)

Run `cosmos init` (with no arguments). Cosmos will prompt you for:

1. **Project name** — directory name (alphanumeric, hyphens, underscores only)
2. **Template** — built-in (api, worker, cli) or external (from GitHub; list from API and `manifest.yaml`)
3. **Module path** — Go module (e.g. `github.com/your-org/myapp`); default suggested from `$USER` and project name
4. **Overwrite?** — if the directory already exists, confirm to replace it

```bash
cosmos init
# or explicitly
cosmos init --interactive
cosmos init -i
```

You choose the template in the menu—built-in (api, worker, cli) or external (from GitHub). Cosmos validates project name and module path as you go.

### Listing templates and packages

- **Built-in + external templates:** `cosmos init --list` or `cosmos init -l`
- **External templates only (from GitHub):** `cosmos list templates`
- **Packages (from GitHub):** `cosmos list pkgs` or `cosmos list packages`

**Cache:** Templates and packages are cached under `~/.cache/cosmos/`: templates at `~/.cache/cosmos/templates/_repo`, packages at `~/.cache/cosmos/packages/_repo`. To refresh (git pull): `cosmos update` or `cosmos cache refresh`. If a cache does not exist yet, nothing is done for it; the first `cosmos init` (with external template) or `cosmos pkg` creates it.

**GitHub API:** Requests use a 30s timeout. Set `GITHUB_TOKEN` for higher rate limits (e.g. in CI).

### Packages

From the root of your Go project (where `go.mod` is):

- `cosmos pkg` — interactive: choose one or more packages to install into `pkg/`.
- `cosmos pkg <name>` — install a single package. Use `--force` to overwrite existing `pkg/<name>`.

List options: `cosmos list pkgs` (or `cosmos list packages`).

## Installation

Install Cosmos using the install script. The script runs all validations (OS, architecture, binary availability, install directory) so that installation can complete successfully.

**Linux / macOS:**

```bash
curl -sSL https://raw.githubusercontent.com/cosmos-toolkit/cosmos-cli/main/scripts/install.sh | sh
```

The script downloads the matching binary from [GitHub Releases](https://github.com/cosmos-toolkit/cosmos-cli/releases) and installs it to `~/.local/bin` or `/usr/local/bin`. If the directory is not in your `PATH`, the script will tell you what to add (e.g. `export PATH="$HOME/.local/bin:$PATH"` in `~/.zshrc` or `~/.bashrc`).

**Windows:** the script does not support Windows; download the `.zip` for your architecture from [Releases](https://github.com/cosmos-toolkit/cosmos-cli/releases), extract `cosmos.exe`, and add it to your `PATH`.

---

The new project is always created in the **current working directory**. Run `cosmos init` from the folder where you want the project (e.g. `cd ~/labs` then `cosmos init`).

## Contributing

**Build** (from the repo root):

```bash
go mod download
make build
# binary: bin/cosmos
```

**Test:** run the CLI locally and confirm the interactive flow and generated project:

```bash
mkdir -p /tmp/cosmos-test && cd /tmp/cosmos-test
../path/to/cosmos-cli/bin/cosmos init
# Choose template (built-in or external), enter project name and module path
ls <project-name>/
cd <project-name> && go build ./...
```

Run the test suite when relevant: `go test ./...`
