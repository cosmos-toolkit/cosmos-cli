package resolver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	packagesRepoURL  = "https://github.com/cosmos-toolkit/packages"
	packagesCacheDir = ".cache/cosmos/packages"
	packagesRepoDir  = "_repo"
)

// ResolvePackagesRepo clones or updates the packages repo with sparse checkout
// for the "pkg" directory and returns the path to the repo root.
func ResolvePackagesRepo() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	baseCache := filepath.Join(homeDir, packagesCacheDir)
	repoPath := filepath.Join(baseCache, packagesRepoDir)
	pkgPath := filepath.Join(repoPath, "pkg")

	if _, err := os.Stat(pkgPath); err == nil {
		// Already have pkg/; try to pull
		if isGitRepo(repoPath) {
			pullCmd := exec.Command("git", "pull")
			pullCmd.Dir = repoPath
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
			_ = pullCmd.Run()
		}
		return repoPath, nil
	}

	if err := os.MkdirAll(baseCache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	if isGitRepo(repoPath) {
		if err := runInDir(repoPath, "git", "sparse-checkout", "add", "pkg"); err != nil {
			return "", fmt.Errorf("failed to add pkg to sparse checkout: %w", err)
		}
		if err := runInDir(repoPath, "git", "pull"); err != nil {
			return "", fmt.Errorf("failed to pull: %w", err)
		}
		return repoPath, nil
	}

	os.RemoveAll(repoPath)
	cloneCmd := exec.Command("git", "clone",
		"--depth", "1",
		"--filter=blob:none",
		"--sparse",
		packagesRepoURL,
		repoPath,
	)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone packages repo: %w", err)
	}

	if err := runInDir(repoPath, "git", "sparse-checkout", "set", "pkg"); err != nil {
		return "", fmt.Errorf("failed to sparse-checkout pkg: %w", err)
	}

	return repoPath, nil
}
