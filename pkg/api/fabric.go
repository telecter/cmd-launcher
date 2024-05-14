package api

import (
	"fmt"

	util "github.com/telecter/cmd-launcher/internal"
)

type FabricVersionList []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}
type FabricLibrary struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Md5    string `json:"md5,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	Sha512 string `json:"sha512,omitempty"`
	Size   int    `json:"size,omitempty"`
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
	Libraries []FabricLibrary `json:"libraries"`
}

const FabricURLPrefix = "https://meta.fabricmc.net/v2/versions/loader"
const QuiltURLPrefix = "https://meta.quiltmc.org/v3/versions/loader"

func GetLoaderMeta(prefix string, gameVersion string) (FabricMeta, error) {
	list := FabricVersionList{}
	meta := FabricMeta{}
	err := util.GetJSON(fmt.Sprintf("%s/%s", prefix, gameVersion), &list)
	if err != nil {
		return meta, fmt.Errorf("Couldn't retrieve Fabric loader list (%s)", err)
	}
	err = util.GetJSON(fmt.Sprintf("%s/%s/%s/profile/json", prefix, gameVersion, list[0].Loader.Version), &meta)
	if err != nil {
		return meta, fmt.Errorf("Couldn't retrieve Fabric metadata for loader version %s (%s)", list[0].Loader.Version, err)
	}
	return meta, nil
}
