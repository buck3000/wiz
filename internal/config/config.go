package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
)

// AgentConfig defines a custom agent command.
type AgentConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// Config holds user-configurable wiz settings.
type Config struct {
	DefaultStrategy string                 `json:"default_strategy"` // auto, worktree, clone
	PromptEmoji     string                 `json:"prompt_emoji"`
	StatusCacheTTL  time.Duration          `json:"-"`
	StatusCacheTTLs string                 `json:"status_cache_ttl"` // e.g. "2s"
	Agents          map[string]AgentConfig `json:"agents,omitempty"`
}

// Defaults returns the default configuration.
func Defaults() Config {
	return Config{
		DefaultStrategy: "auto",
		PromptEmoji:     "\U0001f9d9", // 🧙
		StatusCacheTTL:  2 * time.Second,
		StatusCacheTTLs: "2s",
	}
}

// Load reads config from <wiz-dir>/config.json, returning defaults for missing fields.
func Load(repo *gitx.Repo) Config {
	cfg := Defaults()
	data, err := os.ReadFile(filepath.Join(WizDir(repo), "config.json"))
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.PromptEmoji == "" {
		cfg.PromptEmoji = Defaults().PromptEmoji
	}
	if cfg.StatusCacheTTLs != "" {
		if d, err := time.ParseDuration(cfg.StatusCacheTTLs); err == nil {
			cfg.StatusCacheTTL = d
		}
	}
	return cfg
}
