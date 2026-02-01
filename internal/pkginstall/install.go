// Package pkginstall instala pacotes do repositório cosmos-toolkit/packages
// no projeto atual: copia código, reescreve imports e instala dependências.
package pkginstall

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cosmos-toolkit/cosmos-cli/internal/github"
	"github.com/cosmos-toolkit/cosmos-cli/internal/resolver"
	"gopkg.in/yaml.v3"
)

const pkgsModule = "github.com/cosmos-toolkit/pkgs"

// Manifest descreve os pacotes e suas dependências.
type Manifest struct {
	Packages map[string]PackageMeta `yaml:"packages"`
}

// PackageMeta contém copy_deps (pacotes do repo a copiar) e go_get (deps externas).
type PackageMeta struct {
	CopyDeps []string `yaml:"copy_deps"`
	GoGet    []string `yaml:"go_get"`
}

// Install copia o pacote name e seus copy_deps para pkg/ no cwd, reescreve
// imports para o module do projeto e executa go get para go_get.
func Install(name, cwd string) error {
	manifestData, err := github.GetPackagesManifest()
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	meta, ok := manifest.Packages[name]
	if !ok {
		return fmt.Errorf("package %q not found in manifest", name)
	}

	repoPath, err := resolver.ResolvePackagesRepo()
	if err != nil {
		return fmt.Errorf("failed to resolve packages repo: %w", err)
	}

	srcPkg := filepath.Join(repoPath, "pkg")
	dstPkg := filepath.Join(cwd, "pkg")

	modulePath, err := readModulePath(filepath.Join(cwd, "go.mod"))
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	toCopy := []string{name}
	toCopy = append(toCopy, meta.CopyDeps...)
	seen := make(map[string]bool)
	var ordered []string
	for _, n := range toCopy {
		if seen[n] {
			continue
		}
		seen[n] = true
		ordered = append(ordered, n)
	}

	for _, n := range ordered {
		src := filepath.Join(srcPkg, n)
		dst := filepath.Join(dstPkg, n)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("package %q not found in repo: %w", n, err)
		}
		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("failed to copy %q: %w", n, err)
		}
	}

	if err := rewriteImportsInDir(dstPkg, pkgsModule, modulePath); err != nil {
		return fmt.Errorf("failed to rewrite imports: %w", err)
	}

	for _, imp := range meta.GoGet {
		if err := runGoGet(cwd, imp); err != nil {
			return fmt.Errorf("go get %s: %w", imp, err)
		}
	}

	if err := runGoModTidy(cwd); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	return nil
}

func readModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func rewriteImportsInDir(dir, fromModule, toModule string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newData := strings.ReplaceAll(string(data), fromModule, toModule)
		if newData == string(data) {
			return nil
		}
		return os.WriteFile(path, []byte(newData), info.Mode())
	})
}

func runGoGet(cwd, pkg string) error {
	cmd := exec.Command("go", "get", pkg)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGoModTidy(cwd string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
