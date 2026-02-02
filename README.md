# 🌌 Cosmos

**From chaos to scalable structure.**

Cosmos is a CLI designed to help engineers initialize systems with clarity, structure, and long-term maintainability in mind.

It doesn’t generate folders randomly. It doesn’t hide decisions behind magic. Cosmos organizes systems from the very beginning — so they can grow, scale, and evolve without friction.

## Why Cosmos exists

Every system starts the same way: with uncertainty.

New projects often begin as a collection of guesses:

- unclear boundaries
- inconsistent structure
- premature abstractions
- shortcuts that become permanent

Cosmos exists to turn that initial chaos into **intentional structure**.

Not by enforcing a framework. Not by hiding complexity. But by providing a clear, explicit starting point.

## What Cosmos is?

Cosmos gives:

- a system initializer, not a framework
- explicit by default
- predictable and deterministic
- focused on long-term maintenance
- designed to scale without accidental complexity

## Core concepts

Cosmos works with **systems**, not stacks.

### System types

Cosmos starts from a small set of well-defined system types (built-in and external):

- **api** — HTTP services and public interfaces
- **worker** — background processing and async jobs
- **cli** — developer tools and automation

External templates (e.g. api-hexagonal, monorepo-starter) are listed in the interactive menu. Each type has a clear responsibility and a coherent structure.

**Interactive (recommended):** run `cosmos init` and choose template and options from the prompts.

**Non-interactive:** pass type, name, and flags:

```bash
cosmos init api payments --module github.com/your-org/payments
cosmos init worker jobs --module github.com/your-org/jobs
cosmos init cli toolbox --module github.com/your-org/toolbox
```

## Templates as contracts

In Cosmos, templates are not just file trees. They are contracts.

Each template explicitly declares:

- what it generates
- what it expects
- what decisions are made upfront

This keeps generation predictable and maintainable over time.

## Designed with Go’s philosophy in mind

Cosmos is built around the same principles that guide Go:

- clarity over cleverness
- explicit over implicit
- composition over frameworks
- simplicity that scales

Cosmos is Go-first, but not Go-only.
The ideas apply to systems — not just code.

## Usage

### Interactive mode (recommended)

The simplest way to create a project: run `cosmos init` with no arguments. Cosmos will ask:

1. **Project name** — directory name for the new project
2. **Template** — choose from a list:
   - **Built-in:** api, worker, cli
   - **External (GitHub):** api-clean-arch, api-grpc, api-hexagonal, cli, monorepo-starter, worker-cron, worker-queue
3. **Module path** — Go module (e.g. `github.com/your-org/myapp`); a default is suggested from your username and project name
4. **Overwrite?** — if the directory already exists, confirm to replace it

```bash
cosmos init
# or explicitly
cosmos init --interactive
cosmos init -i
```

No flags needed. Cosmos validates names and module paths as you go.

### Listing templates and packages

Before or instead of the interactive flow, you can list what’s available:

```bash
cosmos list templates   # Built-in + external templates (from GitHub)
cosmos list pkgs        # Reusable packages (logger, config, validator, ...)
```

**Refreshing the cache:** templates and packages are cached under `~/.cache/cosmos/`. To pull the latest versions:

```bash
cosmos update           # git pull in both caches
cosmos cache refresh    # same as above
```

If a cache does not exist yet, nothing is done for it; use `cosmos init` or `cosmos pkg` to create it on first use.

**GitHub API:** requests use a 30s timeout to avoid hanging. If you hit rate limits (e.g. in CI), set `GITHUB_TOKEN` so the CLI uses authenticated requests and gets higher limits.

### Command-line mode (non-interactive)

When you already know the type and flags, you can skip the wizard:

**Built-in types (api, worker, cli):**

```bash
cosmos init api payments --module github.com/your-org/payments
cosmos init worker jobs --module github.com/your-org/jobs
cosmos init cli toolbox --module github.com/your-org/toolbox
```

**External templates (from GitHub):**
Templates are fetched from `github.com/cosmos-toolkit/templates` via Git sparse checkout and cached under `~/.cache/cosmos/templates/`. Descriptions shown in `cosmos list templates` and the interactive init menu come from a root `manifest.yaml` in that repo (same pattern as packages). Example:

```yaml
templates:
  api-hexagonal:
    description: "API with hexagonal architecture"
  monorepo-starter:
    description: "Monorepo with multiple services"
```

```bash
cosmos init myapp --module github.com/your-org/myapp --template api-hexagonal
cosmos init myapp --module github.com/your-org/myapp --template monorepo-starter
```

**Flags:**

- `--module` (required in non-interactive mode) — Go module path
- `--template` — External template name (e.g. `api-clean-arch`, `worker-queue`)
- `--force` — Overwrite existing directory
- `--list`, `-l` — List available templates and exit

### Adding packages to an existing project

From the root of a Go project (where `go.mod` is):

- **Interactive:** run `cosmos pkg` or `cosmos pkg -i` to list packages, select one or more with the arrow keys and space, then confirm with Enter.
- **By name:** pass the package name as argument.

If `pkg/<name>` already exists, the command **fails** unless you use `--force` (or answer "Overwrite?" in interactive mode). Use `--force` / `-f` to overwrite without prompting (useful for automation).

```bash
cosmos pkg              # interactive: list and select packages (single or multiple)
cosmos pkg -i           # same as above
cosmos pkg logger       # copies pkg/logger + copy_deps, rewrites imports
cosmos pkg config      # copies pkg/config, runs go get for dependencies
cosmos pkg validator
cosmos pkg logger --force   # overwrite existing pkg/logger
```

Use `cosmos list pkgs` to see all available packages.

---

**Cosmos will:**

- initialize a clear structure
- apply the selected system type or template
- validate configuration
- generate a maintainable starting point

No magic. No surprises.

## Installation

Cosmos is a single binary. Install it once and run `cosmos` from any directory (like `docker` or `go`).

### Option 1: Install without Go (recommended)

**Linux / macOS** — run the install script (downloads the latest release binary):

```bash
curl -sSL https://raw.githubusercontent.com/cosmos-toolkit/cosmos-cli/main/scripts/install.sh | sh
```

The script detects your OS and architecture, downloads the matching binary from [GitHub Releases](https://github.com/cosmos-toolkit/cosmos-cli/releases), and installs it to `~/.local/bin` (or `/usr/local/bin` if writable). Add that directory to your `PATH` if needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

**Windows** — download the `.zip` for your architecture from [Releases](https://github.com/cosmos-toolkit/cosmos-cli/releases), extract `cosmos.exe`, and place it in a directory that is in your `PATH`.

### Option 2: Install with Go

If you have Go 1.21+ installed:

```bash
go install github.com/cosmos-toolkit/cosmos-cli/cmd/cosmos@latest
```

The binary is placed in `$HOME/go/bin`. Ensure that directory is in your `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Option 3: Install from source (development)

If you have the repo cloned:

```bash
cd /caminho/para/cosmos-cli
make install
```

---

The new project is always created in the **current working directory**. Run `cosmos init` from the folder where you want the project (e.g. `cd ~/labs` then `cosmos init api ...`).

## Building

```bash
make build
# or
go build -o bin/cosmos ./cmd/cosmos
```

## Testing the project

### 1. Build the CLI

```bash
cd /caminho/para/cosmos-cli
go mod download   # baixa dependências (precisa de rede)
make build
```

### 2. Test interactive mode

Crie um diretório temporário e rode o cosmos (sem argumentos) para abrir o menu interativo:

```bash
mkdir -p /tmp/cosmos-test && cd /tmp/cosmos-test
cosmos init
# Responda: project name, escolha um template (api/worker/cli ou externo), module path
ls <nome-do-projeto>/
```

Se o cosmos estiver no `PATH` (por exemplo após `make install`), use `cosmos init` de qualquer pasta.

### 3. Test non-interactive (built-in templates)

```bash
cd /tmp/cosmos-test

# API
cosmos init api payments --module github.com/you/payments
ls payments/                    # deve listar go.mod, cmd/, internal/, README.md, etc.
cd payments && go build ./...   # deve compilar

cd /tmp/cosmos-test
# Worker
cosmos init worker jobs --module github.com/you/jobs
ls jobs/

# CLI
cosmos init cli toolbox --module github.com/you/toolbox
ls toolbox/
```

### 4. Test validation (erros esperados)

```bash
# Falta --module
cosmos init api payments
# Error: --module is required

# Tipo inválido
cosmos init invalid payments --module github.com/you/payments
# Error: invalid type

# Nome do projeto inválido (espaço)
cosmos init api "my project" --module github.com/you/payments
# Error: project name can only contain...
```

### 5. Test with --force

```bash
cosmos init api payments --module github.com/you/payments
cosmos init api payments --module github.com/you/payments
# Error: directory payments already exists. Use --force to overwrite

cosmos init api payments --module github.com/you/payments --force
# Deve sobrescrever sem erro
```

### 6. Test external template (opcional)

Requer rede e o repositório `github.com/cosmos-toolkit/templates/<nome>` existir:

```bash
cosmos init myapp --module github.com/you/myapp --template api-hexagonal
# ou: cosmos list templates  para ver todos os externos
```

## Project structure

Templates are embedded from `cmd/cosmos/templates/` (api, worker, cli). The entry point is `cmd/cosmos/main.go`.
