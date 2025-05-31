package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

var RootDir string
var LibrariesDir string
var InstancesDir string
var CachesDir string
var AssetsDir string
var AuthStorePath string

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
