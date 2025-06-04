package meta

import (
	"fmt"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

// A FabricVersionList is a list of all Fabric loader versions.
type FabricVersionList []struct {
	Separator string `json:"separator"`
	Build     int    `json:"build"`
	Maven     string `json:"maven"`
	Version   string `json:"version"`
	Stable    bool   `json:"stable"`
}

// A FabricMeta is metadata of the libraries and other data needed to start a Fabric-modded game.
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
	Libraries []FabricLibrary
}

type FabricLibrary struct {
	URL  string `json:"url,omitempty"`
	Sha1 string `json:"sha1,omitempty"`
	Size int    `json:"size,omitempty"`
	Name LibrarySpecifier
}

func (library FabricLibrary) Artifact() Artifact {
	path := library.Name.Path()
	return Artifact{
		Path: path,
		URL:  library.URL + "/" + path,
		Sha1: library.Sha1,
		Size: library.Size,
	}
}
func (library FabricLibrary) ShouldInstall() bool {
	return true
}
func (library FabricLibrary) Specifier() LibrarySpecifier {
	return library.Name
}

// FabricLoader represents a variant of the Fabric mod loader.
type FabricLoader string

const (
	FabricLoaderStandard FabricLoader = "fabric" // FabricMC - https://fabricmc.net
	FabricLoaderQuilt    FabricLoader = "quilt"  // Quilt - https://quiltmc.org
)

func (fabricLoader FabricLoader) String() string {
	return string(fabricLoader)
}

const FABRIC_VERSIONS_URL = "https://meta.fabricmc.net/v2/versions/loader"
const FABRIC_PROFILES_URL = "https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json"
const QUILT_VERSIONS_URL = "https://meta.quiltmc.org/v3/versions/loader"
const QUILT_PROFILES_URL = "https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json"

// GetFabricVersions retrieves a list of all versions of the specified Fabric variant.
func GetFabricVersions(fabricLoader FabricLoader) (FabricVersionList, error) {
	cache := network.JSONCache[FabricVersionList]{
		Path: filepath.Join(env.CachesDir, fabricLoader.String(), "versions.json"),
	}
	switch fabricLoader {
	case FabricLoaderStandard:
		cache.URL = FABRIC_VERSIONS_URL
	case FabricLoaderQuilt:
		cache.URL = QUILT_VERSIONS_URL
	}
	var versions FabricVersionList
	if err := cache.UpdateAndRead(&versions); err != nil {
		if err := cache.Read(&versions); err != nil {
			return FabricVersionList{}, err
		}
	}

	return versions, nil
}

// GetFabricMeta retrieves the launcher metadata for a loader version and game version of the specified Fabric variant.
func GetFabricMeta(gameVersion string, loaderVersion string, fabricLoader FabricLoader) (FabricMeta, error) {
	cache := network.JSONCache[FabricMeta]{
		Path: filepath.Join(env.CachesDir, fabricLoader.String(), loaderVersion+"-"+gameVersion+".json"),
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
			return FabricMeta{}, fmt.Errorf("retrieve metadata for %s version %s: %w", fabricLoader, loaderVersion, err)
		}
	}
	return meta, nil
}
