package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	templatesRepo    = "https://api.github.com/repos/cosmos-toolkit/templates/contents"
	templatesRepoURL = "https://github.com/cosmos-toolkit/templates"
	packagesRepo     = "https://api.github.com/repos/cosmos-toolkit/packages/contents"
	packagesRepoURL  = "https://github.com/cosmos-toolkit/packages"
	packagesBranch   = "main"
	templatesBranch  = "main"

	// defaultTimeout is the timeout for GitHub API requests (avoids hanging).
	defaultTimeout = 30 * time.Second
)

// httpClient is used for all GitHub API calls (has timeout; no retry).
var httpClient = &http.Client{Timeout: defaultTimeout}

type contentItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TemplateInfo holds name, description, and link for display.
type TemplateInfo struct {
	Name        string
	Description string
	Link        string
}

// templatesManifest describes manifest.yaml format at templates repo root (like packages).
type templatesManifest struct {
	Templates map[string]struct {
		Description string `yaml:"description"`
	} `yaml:"templates"`
}

// ListTemplates returns template names from cosmos-toolkit/templates.
func ListTemplates() ([]string, error) {
	return listDirs(templatesRepo)
}

// GetTemplatesManifest returns the manifest.yaml content from cosmos-toolkit/templates.
func GetTemplatesManifest() ([]byte, error) {
	return GetFile(templatesRepo, "manifest.yaml")
}

// ListTemplatesWithInfo returns templates with description and link (from manifest if available).
func ListTemplatesWithInfo() ([]TemplateInfo, error) {
	names, err := listDirs(templatesRepo)
	if err != nil {
		return nil, err
	}

	descriptions := make(map[string]string)
	if data, err := GetTemplatesManifest(); err == nil {
		var m templatesManifest
		if yaml.Unmarshal(data, &m) == nil && m.Templates != nil {
			for k, v := range m.Templates {
				descriptions[k] = v.Description
			}
		}
	}

	var templates []TemplateInfo
	for _, name := range names {
		desc := descriptions[name]
		if desc == "" {
			desc = "-"
		}
		templates = append(templates, TemplateInfo{
			Name:        name,
			Description: desc,
			Link:        fmt.Sprintf("%s/tree/%s/%s", templatesRepoURL, templatesBranch, name),
		})
	}
	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	return templates, nil
}

// ListPackages returns package names from cosmos-toolkit/packages (subdir pkg/).
func ListPackages() ([]string, error) {
	return listDirs(packagesRepo + "/pkg")
}

// PackageInfo holds name, description, and link for display.
type PackageInfo struct {
	Name        string
	Description string
	Link        string
}

// packagesManifest describes manifest.yaml format for descriptions.
type packagesManifest struct {
	Packages map[string]struct {
		Description string `yaml:"description"`
	} `yaml:"packages"`
}

// ListPackagesWithInfo returns packages with description and link (from manifest if available).
func ListPackagesWithInfo() ([]PackageInfo, error) {
	names, err := listDirs(packagesRepo + "/pkg")
	if err != nil {
		return nil, err
	}

	descriptions := make(map[string]string)
	if data, err := GetPackagesManifest(); err == nil {
		var m packagesManifest
		if yaml.Unmarshal(data, &m) == nil && m.Packages != nil {
			for k, v := range m.Packages {
				descriptions[k] = v.Description
			}
		}
	}

	var pkgs []PackageInfo
	for _, name := range names {
		desc := descriptions[name]
		if desc == "" {
			desc = "-"
		}
		pkgs = append(pkgs, PackageInfo{
			Name:        name,
			Description: desc,
			Link:        fmt.Sprintf("%s/tree/%s/pkg/%s", packagesRepoURL, packagesBranch, name),
		})
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}

// doRequest performs a GET to url with context timeout and optional GITHUB_TOKEN.
func doRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return httpClient.Do(req)
}

func listDirs(apiURL string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := doRequest(ctx, apiURL)
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := doRequest(ctx, url)
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
