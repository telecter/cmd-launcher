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
	cache := network.Cache[FabricVersionList]{
		Path:        filepath.Join(env.CachesDir, f.name, "versions.json"),
		URL:         f.versions,
		AlwaysFetch: true,
	}
	var versions FabricVersionList
	if err := cache.Read(&versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// FetchMeta retrieves version metadata for the specified game and loader version of Fabric.
//
// Besides normal version identifiers, loaderVersion can also be "latest".
func (f fabricAPI) FetchMeta(gameVersion, loaderVersion string) (VersionMeta, error) {
	if loaderVersion == "latest" {
		versions, err := f.FetchVersions()
		if err != nil {
			return VersionMeta{}, fmt.Errorf("fetch versions: %w", err)
		}
		loaderVersion = versions[0].Version
	}
	cache := network.Cache[VersionMeta]{
		Path: filepath.Join(env.CachesDir, f.name, loaderVersion+"-"+gameVersion+".json"),
		URL:  fmt.Sprintf(f.profiles, gameVersion, loaderVersion),
	}

	var fabricMeta VersionMeta
	if err := cache.Read(&fabricMeta); err != nil {
		var statusErr *network.HTTPStatusError
		if errors.As(err, &statusErr) {
			if statusErr.StatusCode == 400 || statusErr.StatusCode == 404 {
				return VersionMeta{}, fmt.Errorf("invalid or unsuitable game/fabric version")
			}
		}
		return VersionMeta{}, err
	}
	fabricMeta.LoaderID = loaderVersion
	return fabricMeta, nil
}
