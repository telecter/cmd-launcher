package meta

import (
	"fmt"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/network"
)

type FabricVersionList []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}
type FabricMeta struct {
	ID           string `json:"id"`
	InheritsFrom string `json:"inheritsFrom"`
	ReleaseTime  string `json:"releaseTime"`
	Time         string `json:"time"`
	Type         string `json:"type"`
	MainClass    string `json:"mainClass"`
	Arguments    struct {
		Game []any    `json:"game"`
		Jvm  []string `json:"jvm"`
	} `json:"arguments"`
	Libraries []Library
}

func GetFabricVersions(gameVersion string) (FabricVersionList, error) {
	cache := network.JSONCache[FabricVersionList]{
		Path: filepath.Join(internal.CachesDir, "fabric", gameVersion+"-versions.json"),
		URL:  fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s", gameVersion),
	}
	var versions FabricVersionList
	if err := cache.UpdateAndRead(&versions); err != nil {
		if err := cache.Read(&versions); err != nil {
			return FabricVersionList{}, fmt.Errorf("retrieve Fabric versions: %w", err)
		}
	}

	return versions, nil
}

func GetQuiltVersions(gameVersion string) (FabricVersionList, error) {
	cache := network.JSONCache[FabricVersionList]{
		Path: filepath.Join(internal.CachesDir, "quilt", gameVersion+"-versions.json"),
		URL:  fmt.Sprintf("https://meta.quiltmc.org/v3/versions/loader/%s", gameVersion),
	}
	var versions FabricVersionList
	if err := cache.UpdateAndRead(&versions); err != nil {
		if err := cache.Read(&versions); err != nil {
			return FabricVersionList{}, fmt.Errorf("retrieve Quilt versions: %w", err)
		}
	}

	return versions, nil
}

func GetFabricMeta(gameVersion string, loaderVersion string) (FabricMeta, error) {
	cache := network.JSONCache[FabricMeta]{
		Path: filepath.Join(internal.CachesDir, "fabric", loaderVersion+"-"+gameVersion+".json"),
		URL:  fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", gameVersion, loaderVersion),
	}
	var meta FabricMeta
	if err := cache.Read(&meta); err != nil {
		if err := cache.UpdateAndRead(&meta); err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve metadata for Fabric version %s: %w", loaderVersion, err)
		}
	}
	return meta, nil
}

func GetQuiltMeta(gameVersion string, loaderVersion string) (FabricMeta, error) {
	cache := network.JSONCache[FabricMeta]{
		Path: filepath.Join(internal.CachesDir, "quilt", loaderVersion+"-"+gameVersion+".json"),
		URL:  fmt.Sprintf("https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json", gameVersion, loaderVersion),
	}
	var meta FabricMeta
	if err := cache.Read(&meta); err != nil {
		if err := cache.UpdateAndRead(&meta); err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve metadata for Quilt version %s: %w", loaderVersion, err)
		}
	}

	return meta, nil
}
