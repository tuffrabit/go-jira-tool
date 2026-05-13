// Copyright 2026 Robert Boucher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	BaseURL        string `json:"base_url"`
	Email          string `json:"email"`
	APIToken       string `json:"api_token"`
	DefaultProject string `json:"default_project"`
}

// AppDir returns the path to ~/.go-jira-tool, creating it if necessary.
func AppDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".go-jira-tool")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}
	return dir, nil
}

// IssuesDir returns the path to ~/.go-jira-tool/issues, creating it if necessary.
func IssuesDir() (string, error) {
	appDir, err := AppDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appDir, "issues")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create issues directory: %w", err)
	}
	return dir, nil
}

func LoadConfig() (*Config, error) {
	// Try ~/.go-jira-tool/config.json first, then fall back to next to executable
	configPath := ""

	appDir, err := AppDir()
	if err == nil {
		candidate := filepath.Join(appDir, "config.json")
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
		}
	}

	if configPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to determine executable path: %w", err)
		}
		configPath = filepath.Join(filepath.Dir(exePath), "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (checked ~/.go-jira-tool/config.json and next to executable): %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required in config.json")
	}
	if cfg.Email == "" {
		return nil, fmt.Errorf("email is required in config.json")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("api_token is required in config.json")
	}

	return &cfg, nil
}
