// Package env provides directories used by the launcher for various data.
package env

import (
	"fmt"
	"os"
	"path/filepath"
)

// Base launcher directory. Defaults to "$HOME/.minecraft"
var RootDir string

// Java libraries directory
var LibrariesDir string

// Instances directory
var InstancesDir string

// Caches directory, e.g. version metadata, version manifest.
var CachesDir string

// Game assets directory and asset index
var AssetsDir string

// Path of the global authentication store
var AuthStorePath string

// Set all directories from a root directory. These values can also be changed individually.
// However, they should not be changed between operations, as the launcher will not be able to find necessary files.
func SetDirs(rootDir string) error {
	RootDir = rootDir
	InstancesDir = filepath.Join(RootDir, "instances")
	LibrariesDir = filepath.Join(RootDir, "libraries")
	CachesDir = filepath.Join(RootDir, "caches")
	AssetsDir = filepath.Join(RootDir, "assets")
	AuthStorePath = filepath.Join(RootDir, "account.json")

	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("create root directory: %w", err)
	}
	return nil
}

func init() {
	home, _ := os.UserHomeDir()
	SetDirs(filepath.Join(home, ".minecraft"))
}
