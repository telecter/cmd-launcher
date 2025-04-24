package meta

import (
	"fmt"

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

const fabricVersionsListCache = "fabric/%s-versions.json"
const fabricLauncherProfileCache = "fabric/%s-profile.json"

func GetFabricMeta(gameVersion string) (FabricMeta, error) {
	versionsCache := fmt.Sprintf(fabricVersionsListCache, gameVersion)
	profileCache := fmt.Sprintf(fabricLauncherProfileCache, gameVersion)

	var list FabricVersionList
	var err error
	if isCacheValid(versionsCache) {
		err = readCache(versionsCache, &list)
	} else {
		err = network.FetchJSONData(fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s", gameVersion), &list)
		writeCache(versionsCache, list)
	}
	if err != nil {
		return FabricMeta{}, fmt.Errorf("failed to retrieve Fabric loader list: %w", err)
	}

	var meta FabricMeta
	if isCacheValid(profileCache) {
		err = readCache(profileCache, &meta)
	} else {
		err = network.FetchJSONData(fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", gameVersion, list[0].Loader.Version), &meta)
		writeCache(profileCache, meta)
	}
	if err != nil {
		return FabricMeta{}, fmt.Errorf("failed to retrieve Fabric metadata for loader version %s: %w", list[0].Loader.Version, err)
	}
	return meta, nil
}
