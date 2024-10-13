package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	util "codeberg.org/telecter/cmd-launcher/internal"
)

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []Version `json:"versions"`
}
type Version struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`
}
type Artifact struct {
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}
type Library struct {
	Downloads struct {
		Artifact struct {
			Path string `json:"path"`
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"artifact"`
		Classifiers struct {
			NativesMacOS   Artifact `json:"natives-macos"`
			NativesLinux   Artifact `json:"natives-linux"`
			NativesWindows Artifact `json:"natives-windows"`
		} `json:"classifiers"`
	} `json:"downloads"`
	Name    string `json:"name"`
	Natives struct {
		Linux   string `json:"linux"`
		MacOS   string `json:"macos"`
		Windows string `json:"windows"`
	}
	URL   string `json:"url"`
	Rules []struct {
		Action string `json:"action"`
		Os     struct {
			Name string `json:"name"`
		} `json:"os"`
	} `json:"rules,omitempty"`
}
type VersionMeta struct {
	Arguments struct {
		Game []any `json:"game"`
		Jvm  []any `json:"jvm"`
	} `json:"arguments"`
	AssetIndex struct {
		ID        string `json:"id"`
		Sha1      string `json:"sha1"`
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		URL       string `json:"url"`
	} `json:"assetIndex"`
	Assets          string `json:"assets"`
	ComplianceLevel int    `json:"complianceLevel"`
	Downloads       struct {
		Client struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"client"`
		ClientMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"client_mappings"`
		Server struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"server"`
		ServerMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"server_mappings"`
	} `json:"downloads"`
	ID          string `json:"id"`
	JavaVersion struct {
		Component    string `json:"component"`
		MajorVersion int    `json:"majorVersion"`
	} `json:"javaVersion"`
	Libraries []Library `json:"libraries"`
	Logging   struct {
		Client struct {
			Argument string `json:"argument"`
			File     struct {
				ID   string `json:"id"`
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"file"`
			Type string `json:"type"`
		} `json:"client"`
	} `json:"logging"`
	MainClass              string    `json:"mainClass"`
	MinimumLauncherVersion int       `json:"minimumLauncherVersion"`
	ReleaseTime            time.Time `json:"releaseTime"`
	Time                   time.Time `json:"time"`
	Type                   string    `json:"type"`
}
type AssetIndex struct {
	Objects map[string]struct {
		Hash string
		Size int
	}
}

func GetVersionMeta(id string) (VersionMeta, error) {
	manifest := VersionManifest{}
	meta := VersionMeta{}
	err := util.GetJSON("https://launchermeta.mojang.com/mc/game/version_manifest.json", &manifest)

	if err != nil {
		return meta, fmt.Errorf("failed to retrieve version manifest (%s)", err)
	}

	if id == "" {
		id = manifest.Latest.Release
	}

	for _, v := range manifest.Versions {
		if v.ID == id {
			resp, err := http.Get(v.URL)

			if err := util.CheckResponse(resp, err); err != nil {
				return meta, fmt.Errorf("failed to retrieve version metadata (%s)", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			json.Unmarshal(body, &meta)

			return meta, nil
		}
	}
	return meta, fmt.Errorf("invalid version")
}
