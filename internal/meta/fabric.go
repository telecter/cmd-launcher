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

type FabricLoader string

const (
	FabricLoaderStandard FabricLoader = "fabric"
	FabricLoaderQuilt    FabricLoader = "quilt"
)

func (fabricLoader FabricLoader) String() string {
	return string(fabricLoader)
}

const FABRIC_VERSIONS_URL = "https://meta.fabricmc.net/v2/versions/loader/%s"
const FABRIC_PROFILES_URL = "https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json"
const QUILT_VERSIONS_URL = "https://meta.quiltmc.org/v3/versions/loader/%s"
const QUILT_PROFILES_URL = "https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json"

func GetFabricVersions(gameVersion string, fabricLoader FabricLoader) (FabricVersionList, error) {
	cache := network.JSONCache[FabricVersionList]{
		Path: filepath.Join(internal.CachesDir, fabricLoader.String(), gameVersion+"-versions.json"),
	}
	switch fabricLoader {
	case FabricLoaderStandard:
		cache.URL = fmt.Sprintf(FABRIC_VERSIONS_URL, gameVersion)
	case FabricLoaderQuilt:
		cache.URL = fmt.Sprintf(QUILT_VERSIONS_URL, gameVersion)
	}
	var versions FabricVersionList
	if err := cache.UpdateAndRead(&versions); err != nil {
		if err := cache.Read(&versions); err != nil {
			return FabricVersionList{}, fmt.Errorf("retrieve Fabric/Quilt versions: %w", err)
		}
	}

	return versions, nil
}

func GetFabricMeta(gameVersion string, loaderVersion string, fabricLoader FabricLoader) (FabricMeta, error) {
	cache := network.JSONCache[FabricMeta]{
		Path: filepath.Join(internal.CachesDir, "fabric", loaderVersion+"-"+gameVersion+".json"),
	}
	switch fabricLoader {
	case FabricLoaderStandard:
		cache.URL = fmt.Sprintf(FABRIC_PROFILES_URL, gameVersion, loaderVersion)
	case FabricLoaderQuilt:
		cache.URL = fmt.Sprintf(QUILT_PROFILES_URL, gameVersion, loaderVersion)
	}
	var meta FabricMeta
	if err := cache.Read(&meta); err != nil {
		if err := cache.UpdateAndRead(&meta); err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve metadata for Fabric/Quilt version %s: %w", loaderVersion, err)
		}
	}
	return meta, nil
}
