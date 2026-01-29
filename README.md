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

Cosmos starts from a small set of well-defined system types:

- **api** — HTTP services and public interfaces
- **worker** — background processing and async jobs
- **cli** — developer tools and automation

Each type has a clear responsibility and a coherent structure.

```bash
cosmos init api payments
cosmos init worker jobs
cosmos init cli toolbox
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

### Built-in templates (api, worker, cli)

Initialize a project with a built-in type:

```bash
cosmos init api payments --module github.com/your-org/payments
cosmos init worker jobs --module github.com/your-org/jobs
cosmos init cli toolbox --module github.com/your-org/toolbox
```

### External templates (GitHub)

Initialize a project with an external template by name. Templates are fetched from `github.com/cosmos-cli/templates/<name>` and cached under `~/.cache/cosmos/templates/`:

```bash
cosmos init myapp --module github.com/your-org/myapp --template hexagonal-architecture
cosmos init myapp --module github.com/your-org/myapp --template ddd-architecture
```

### Flags

- `--module` (required) — Go module path (e.g. `github.com/user/repo`)
- `--force` — Overwrite existing directory
- `--template` — External template name (fetched from GitHub)

**Cosmos will:**

- initialize a clear structure
- apply the selected system type or template
- validate configuration
- generate a maintainable starting point

No magic. No surprises.

## Building

```bash
make build
# or
go build -o bin/cosmos ./cmd/cosmos
```

## Testing the project

### 1. Build the CLI

```bash
cd /caminho/para/cli
go mod download   # baixa dependências (precisa de rede)
make build
```

### 2. Test with built-in templates

Crie um diretório temporário e rode o cosmos a partir do binário:

```bash
mkdir -p /tmp/cosmos-test && cd /tmp/cosmos-test

# API
/path/to/cli/bin/cosmos init api payments --module github.com/you/payments
ls payments/                    # deve listar go.mod, cmd/, internal/, README.md, etc.
cd payments && go build ./...   # deve compilar

cd /tmp/cosmos-test

# Worker
/path/to/cli/bin/cosmos init worker jobs --module github.com/you/jobs
ls jobs/

# CLI
/path/to/cli/bin/cosmos init cli toolbox --module github.com/you/toolbox
ls toolbox/
```

Se o cosmos estiver no `PATH` (por exemplo após `make install`):

```bash
cosmos init api payments --module github.com/you/payments
cosmos init worker jobs --module github.com/you/jobs
cosmos init cli toolbox --module github.com/you/toolbox
```

### 3. Test validation (erros esperados)

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

### 4. Test with --force

```bash
cosmos init api payments --module github.com/you/payments
cosmos init api payments --module github.com/you/payments
# Error: directory payments already exists. Use --force to overwrite

cosmos init api payments --module github.com/you/payments --force
# Deve sobrescrever sem erro
```

### 5. Test external template (opcional)

Requer rede e o repositório `github.com/cosmos-cli/templates/<nome>` existir:

```bash
cosmos init myapp --module github.com/you/myapp --template hexagonal-architecture
```

## Project structure

Templates are embedded from `cmd/cosmos/templates/` (api, worker, cli). The entry point is `cmd/cosmos/main.go`.
