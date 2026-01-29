package resolver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	githubBaseURL = "github.com/cosmos-cli/templates"
	cacheDir      = ".cache/cosmos/templates"
)

func Resolve(templateName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cachePath := filepath.Join(homeDir, cacheDir, templateName)

	// Check if template exists in cache
	if _, err := os.Stat(cachePath); err == nil {
		// Verify it has template.yaml
		templateYAML := filepath.Join(cachePath, "template.yaml")
		if _, err := os.Stat(templateYAML); err == nil {
			return cachePath, nil
		}
		// Cache exists but no template.yaml, remove and re-clone
		os.RemoveAll(cachePath)
	}

	// Clone template
	repoURL := fmt.Sprintf("https://%s/%s", githubBaseURL, templateName)
	if err := cloneTemplate(repoURL, cachePath); err != nil {
		return "", fmt.Errorf("failed to clone template: %w", err)
	}

	// Verify template.yaml exists
	templateYAML := filepath.Join(cachePath, "template.yaml")
	if _, err := os.Stat(templateYAML); err != nil {
		return "", fmt.Errorf("template.yaml not found in template")
	}

	return cachePath, nil
}

func cloneTemplate(repoURL, destPath string) error {
	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", repoURL, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
