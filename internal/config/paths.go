package config

import (
	"os"
	"path/filepath"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
)

// WizDir returns <git-common-dir>/wiz/ — the root of all wiz state for this repo.
func WizDir(repo *gitx.Repo) string {
	return filepath.Join(repo.CommonDir, "wiz")
}

// StateFile returns <wiz-dir>/state.json.
func StateFile(repo *gitx.Repo) string {
	return filepath.Join(WizDir(repo), "state.json")
}

// LockFilePath returns <wiz-dir>/wiz.lock.
func LockFilePath(repo *gitx.Repo) string {
	return filepath.Join(WizDir(repo), "wiz.lock")
}

// TreesDir returns <wiz-dir>/trees/ — where worktree-backed contexts live.
func TreesDir(repo *gitx.Repo) string {
	return filepath.Join(WizDir(repo), "trees")
}

// ClonesDir returns <wiz-dir>/clones/ — where clone-backed contexts live.
func ClonesDir(repo *gitx.Repo) string {
	return filepath.Join(WizDir(repo), "clones")
}

// CacheDir returns <wiz-dir>/cache/.
func CacheDir(repo *gitx.Repo) string {
	return filepath.Join(WizDir(repo), "cache")
}

// LicenseFilePath returns ~/.config/wiz/license.json (global, not per-repo).
func LicenseFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "wiz", "license.json")
}
