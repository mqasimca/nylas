// Package plugin provides plugin discovery and execution for the Nylas CLI.
// Follows kubectl-style plugin pattern where plugins are named "nylas-{name}"
// and discovered in PATH.
package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pluginPrefix = "nylas-"

// Plugin represents a discovered plugin.
type Plugin struct {
	Name string // Plugin name (without "nylas-" prefix)
	Path string // Full path to the plugin executable
}

// Discover finds all available plugins in PATH and ~/.nylas/plugins/.
func Discover() ([]Plugin, error) {
	var plugins []Plugin

	// 1. Search in PATH
	pathPlugins, err := discoverInPath()
	if err != nil {
		return nil, fmt.Errorf("failed to discover plugins in PATH: %w", err)
	}
	plugins = append(plugins, pathPlugins...)

	// 2. Search in ~/.nylas/plugins/
	homePlugins, err := discoverInHome()
	if err != nil {
		// Ignore errors from home directory (might not exist)
		// but continue with PATH plugins
	} else {
		plugins = append(plugins, homePlugins...)
	}

	// 3. Deduplicate (PATH takes precedence over ~/.nylas/plugins/)
	plugins = deduplicate(plugins)

	return plugins, nil
}

// discoverInPath finds plugins in the system PATH.
func discoverInPath() ([]Plugin, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, nil
	}

	var plugins []Plugin
	pathDirs := filepath.SplitList(pathEnv)

	for _, dir := range pathDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Skip directories we can't read
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, pluginPrefix) {
				continue
			}

			fullPath := filepath.Join(dir, name)

			// Check if executable
			if !isExecutable(fullPath) {
				continue
			}

			pluginName := strings.TrimPrefix(name, pluginPrefix)
			plugins = append(plugins, Plugin{
				Name: pluginName,
				Path: fullPath,
			})
		}
	}

	return plugins, nil
}

// discoverInHome finds plugins in ~/.nylas/plugins/.
func discoverInHome() ([]Plugin, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	pluginDir := filepath.Join(homeDir, ".nylas", "plugins")

	// Check if directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, pluginPrefix) {
			continue
		}

		fullPath := filepath.Join(pluginDir, name)

		// Check if executable
		if !isExecutable(fullPath) {
			continue
		}

		pluginName := strings.TrimPrefix(name, pluginPrefix)
		plugins = append(plugins, Plugin{
			Name: pluginName,
			Path: fullPath,
		})
	}

	return plugins, nil
}

// Find finds a specific plugin by name.
func Find(name string) (*Plugin, error) {
	plugins, err := Discover()
	if err != nil {
		return nil, err
	}

	for _, plugin := range plugins {
		if plugin.Name == name {
			return &plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin %q not found", name)
}

// deduplicate removes duplicate plugins, keeping the first occurrence.
func deduplicate(plugins []Plugin) []Plugin {
	seen := make(map[string]bool)
	var result []Plugin

	for _, plugin := range plugins {
		if !seen[plugin.Name] {
			seen[plugin.Name] = true
			result = append(result, plugin)
		}
	}

	return result
}

// isExecutable checks if a file is executable.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if file is executable (Unix permission check)
	mode := info.Mode()
	return mode.IsRegular() && (mode.Perm()&0111 != 0)
}

// Validate checks if a plugin is valid by attempting to execute it with --help.
func Validate(plugin *Plugin) error {
	cmd := exec.Command(plugin.Path, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}
	return nil
}
