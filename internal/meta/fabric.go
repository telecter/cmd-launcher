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
	name string
	url  string
}

var Fabric = fabricAPI{
	name: "fabric",
	url:  "https://meta.fabricmc.net/v2",
}
var Quilt = fabricAPI{
	name: "quilt",
	url:  "https://meta.quiltmc.org/v3",
}

// FetchVersions retrieves a list of all versions of Fabric.
func (api fabricAPI) FetchVersions() (FabricVersionList, error) {
	cache := network.Cache[FabricVersionList]{
		Path:        filepath.Join(env.CachesDir, api.name, "versions.json"),
		URL:         fmt.Sprintf("%s/versions/loader", api.url),
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
func (api fabricAPI) FetchMeta(gameVersion, loaderVersion string) (VersionMeta, error) {
	if loaderVersion == "latest" {
		versions, err := api.FetchVersions()
		if err != nil {
			return VersionMeta{}, fmt.Errorf("fetch versions: %w", err)
		}
		loaderVersion = versions[0].Version
	}
	cache := network.Cache[VersionMeta]{
		Path: filepath.Join(env.CachesDir, api.name, loaderVersion+"-"+gameVersion+".json"),
		URL:  fmt.Sprintf("%s/versions/loader/%s/%s/profile/json", api.url, gameVersion, loaderVersion),
	}

	var fabricMeta VersionMeta
	if err := cache.Read(&fabricMeta); err != nil {
		var statusErr *network.HTTPStatusError
		if errors.As(err, &statusErr) && (statusErr.StatusCode == 400 || statusErr.StatusCode == 404) {
			return VersionMeta{}, fmt.Errorf("invalid or unsuitable game/fabric version")
		}
		return VersionMeta{}, err
	}
	fabricMeta.LoaderID = loaderVersion
	return fabricMeta, nil
}
