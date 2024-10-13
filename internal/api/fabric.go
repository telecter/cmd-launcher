package api

import (
	"fmt"

	util "codeberg.org/telecter/cmd-launcher/internal"
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

const FabricURLPrefix = "https://meta.fabricmc.net/v2/versions/loader"
const QuiltURLPrefix = "https://meta.quiltmc.org/v3/versions/loader"

func GetLoaderMeta(prefix string, gameVersion string) (FabricMeta, error) {
	list := FabricVersionList{}
	meta := FabricMeta{}
	err := util.GetJSON(fmt.Sprintf("%s/%s", prefix, gameVersion), &list)
	if err != nil {
		return meta, fmt.Errorf("couldn't retrieve Fabric loader list (%s)", err)
	}
	err = util.GetJSON(fmt.Sprintf("%s/%s/%s/profile/json", prefix, gameVersion, list[0].Loader.Version), &meta)
	if err != nil {
		return meta, fmt.Errorf("couldn't retrieve Fabric metadata for loader version %s (%s)", list[0].Loader.Version, err)
	}
	return meta, nil
}
