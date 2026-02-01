package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	templatesRepo = "https://api.github.com/repos/cosmos-toolkit/templates/contents"
	packagesRepo  = "https://api.github.com/repos/cosmos-toolkit/packages/contents"
)

type contentItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListTemplates returns template names from cosmos-toolkit/templates.
func ListTemplates() ([]string, error) {
	return listDirs(templatesRepo)
}

// ListPackages returns package names from cosmos-toolkit/packages (subdir pkg/).
func ListPackages() ([]string, error) {
	return listDirs(packagesRepo + "/pkg")
}

func listDirs(apiURL string) ([]string, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var items []contentItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var dirs []string
	for _, item := range items {
		if item.Type == "dir" && item.Name != "." && item.Name != ".." {
			dirs = append(dirs, item.Name)
		}
	}
	return dirs, nil
}

// FileContent holds a single file from GitHub contents API.
type FileContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// GetFile returns the raw file content from a repo (path relative to repo root).
func GetFile(repoContentsBaseURL, path string) ([]byte, error) {
	url := repoContentsBaseURL + "/" + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for %s", resp.StatusCode, path)
	}

	var fc FileContent
	if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if fc.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(fc.Content)
	}
	return []byte(fc.Content), nil
}

// GetPackagesManifest returns the manifest.yaml content from cosmos-toolkit/packages.
func GetPackagesManifest() ([]byte, error) {
	return GetFile(packagesRepo, "manifest.yaml")
}
