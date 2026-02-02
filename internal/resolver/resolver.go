package resolver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	repoURL  = "https://github.com/cosmos-toolkit/templates"
	cacheDir = ".cache/cosmos/templates"
	repoDir  = "_repo"
)

func Resolve(templateName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	baseCache := filepath.Join(homeDir, cacheDir)
	repoPath := filepath.Join(baseCache, repoDir)
	templatePath := filepath.Join(repoPath, templateName)
	templateYAML := filepath.Join(templatePath, "template.yaml")

	// Already in cache with valid template.yaml
	if _, err := os.Stat(templateYAML); err == nil {
		return templatePath, nil
	}

	// Ensure cache base exists
	if err := os.MkdirAll(baseCache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	repoExists := isGitRepo(repoPath)

	if repoExists {
		// Repo exists: add this template to sparse checkout and pull
		if err := addTemplateToSparseCheckout(repoPath, templateName); err != nil {
			return "", fmt.Errorf("failed to add template to sparse checkout: %w", err)
		}
	} else {
		// Fresh clone with sparse checkout for this template only
		if err := cloneWithSparseCheckout(repoPath, templateName); err != nil {
			return "", fmt.Errorf("failed to clone template: %w", err)
		}
	}

	// template.yaml is required; if missing, create a minimal one for compatibility
	if _, err := os.Stat(templateYAML); err != nil {
		if err := writeMinimalTemplateYAML(templateYAML, templateName); err != nil {
			return "", fmt.Errorf("template.yaml not found in template %q and failed to create default: %w", templateName, err)
		}
	}

	return templatePath, nil
}

func writeMinimalTemplateYAML(path, templateName string) error {
	content := fmt.Sprintf(`name: %s
version: "0.1.0"
types: ["%s"]
defaults:
  goVersion: "1.23"
prompts:
  - key: module
    description: "Go module path"
    required: true
  - key: projectName
    description: "Project name"
    required: true
features: []
files:
  engine: gotmpl
  modulePlaceholder: "github.com/your-org/your-app"
`, templateName, templateName)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func cloneWithSparseCheckout(repoPath, templateName string) error {
	parent := filepath.Dir(repoPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	// Remove partial/corrupt clone if exists
	os.RemoveAll(repoPath)

	// Clone with sparse checkout (fetches only the chosen folder)
	cloneCmd := exec.Command("git", "clone",
		"--depth", "1",
		"--filter=blob:none",
		"--sparse",
		repoURL,
		repoPath,
	)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	// Checkout only the template folder (cone mode includes the dir and its contents)
	return runInDir(repoPath, "git", "sparse-checkout", "set", templateName)
}

func addTemplateToSparseCheckout(repoPath, templateName string) error {
	// Add template folder to sparse checkout
	addCmd := exec.Command("git", "sparse-checkout", "add", templateName)
	addCmd.Dir = repoPath
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git sparse-checkout add: %w", err)
	}

	// Pull latest
	pullCmd := exec.Command("git", "pull")
	pullCmd.Dir = repoPath
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}

	return nil
}

func runInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TemplatesRepoPath returns the path to the cached templates repo (~/.cache/cosmos/templates/_repo).
func TemplatesRepoPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, cacheDir, repoDir), nil
}

// PullTemplatesRepo runs git pull in the templates cache if the repo exists.
// It returns (true, nil) when pull ran, (false, nil) when no cache exists, or (_, err) on failure.
func PullTemplatesRepo() (updated bool, err error) {
	repoPath, err := TemplatesRepoPath()
	if err != nil {
		return false, err
	}
	if !isGitRepo(repoPath) {
		return false, nil
	}
	pullCmd := exec.Command("git", "pull")
	pullCmd.Dir = repoPath
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return false, fmt.Errorf("templates: git pull: %w", err)
	}
	return true, nil
}
