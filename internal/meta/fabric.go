package meta

import (
	"errors"
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

type fabricAPI struct {
	name     string // Name of the fabric API.
	versions string // Loader versions endpoint.
	profiles string // Launcher profiles endpoint.
}

var Fabric = fabricAPI{
	name:     "fabric",
	versions: "https://meta.fabricmc.net/v2/versions/loader",
	profiles: "https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json",
}
var Quilt = fabricAPI{
	name:     "quilt",
	versions: "https://meta.quiltmc.org/v3/versions/loader",
	profiles: "https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json",
}

// FetchFabricVersions retrieves a list of all versions of Fabric.
func (f fabricAPI) FetchVersions() (FabricVersionList, error) {
	cache := network.JSONCache[FabricVersionList]{
		Path: filepath.Join(env.CachesDir, f.name, "versions.json"),
		URL:  f.versions,
	}
	var versions FabricVersionList
	if err := cache.FetchAndRead(&versions); err != nil {
		if err := cache.Read(&versions); err != nil {
			return nil, err
		}
	}

	return versions, nil
}

// FetchMeta retrieves version metadata for the specified game and loader version of Fabric.
func (f fabricAPI) FetchMeta(gameVersion, loaderVersion string) (VersionMeta, error) {
	if loaderVersion == "latest" {
		versions, err := f.FetchVersions()
		if err != nil {
			return VersionMeta{}, fmt.Errorf("fetch fabric versions: %w", err)
		}
		loaderVersion = versions[0].Version
	}
	cache := network.JSONCache[VersionMeta]{
		Path: filepath.Join(env.CachesDir, "fabric", loaderVersion+"-"+gameVersion+".json"),
		URL:  fmt.Sprintf(f.profiles, gameVersion, loaderVersion),
	}
	var fabricMeta VersionMeta
	if err := cache.Read(&fabricMeta); err != nil {
		if err := cache.FetchAndRead(&fabricMeta); err != nil {
			var statusErr *network.HTTPStatusError
			if errors.As(err, &statusErr) {
				if statusErr.StatusCode == 400 {
					return VersionMeta{}, fmt.Errorf("invalid fabric version or invalid game version for fabric")
				}
			}
			return VersionMeta{}, err
		}
	}
	fabricMeta.LoaderID = loaderVersion
	return fabricMeta, nil
}
