package internal

import "path/filepath"

var RootDir string
var LibrariesDir string
var InstancesDir string
var CachesDir string
var AssetsDir string
var AccountDataCache string

func SetDirsFromRoot(rootDir string) {
	RootDir = rootDir
	InstancesDir = filepath.Join(RootDir, "instances")
	LibrariesDir = filepath.Join(RootDir, "libraries")
	CachesDir = filepath.Join(RootDir, "caches")
	AssetsDir = filepath.Join(RootDir, "assets")
	AccountDataCache = filepath.Join(RootDir, "account.json")
}
