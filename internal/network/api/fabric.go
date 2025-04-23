package api

import (
	"fmt"
	"log"
	"net/url"

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

const baseEndpoint = "https://meta.fabricmc.net/v2/versions/loader"

func GetLoaderMeta(gameVersion string) (FabricMeta, error) {
	var list FabricVersionList

	loc, _ := url.JoinPath(baseEndpoint, gameVersion)
	err := network.FetchJSONData(loc, &list)
	if err != nil {
		return FabricMeta{}, fmt.Errorf("failed to retrieve Fabric loader list: %w", err)
	}

	var meta FabricMeta

	loc, _ = url.JoinPath(baseEndpoint, gameVersion, list[0].Loader.Version, "profile", "json")
	err = network.FetchJSONData(loc, &meta)
	if err != nil {
		return FabricMeta{}, fmt.Errorf("failed to retrieve Fabric metadata for loader version %s: %w", list[0].Loader.Version, err)
	}

	log.Println(loc)
	return meta, nil
}
