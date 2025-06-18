package meta

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

const (
	VersionManifestURL    = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
	JavaRuntimesURL       = "https://piston-meta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json"
	MinecraftResourcesURL = "https://resources.download.minecraft.net"
	MinecraftLibrariesURL = "https://libraries.minecraft.net"
)

// A VersionManifest is a list of all Minecraft versions.
type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID              string    `json:"id"`
		Type            string    `json:"type"`
		URL             string    `json:"url"`
		Time            time.Time `json:"time"`
		ReleaseTime     time.Time `json:"releaseTime"`
		Sha1            string    `json:"sha1"`
		ComplianceLevel int       `json:"complianceLevel"`
	} `json:"versions"`
}

type JavaManifestList map[string]map[string][]struct {
	Availability struct {
		Group    int `json:"group"`
		Progress int `json:"progress"`
	} `json:"availability"`
	Manifest struct {
		Sha1 string `json:"sha1"`
		Size int    `json:"size"`
		URL  string `json:"url"`
	} `json:"manifest"`
	Version struct {
		Name     string    `json:"name"`
		Released time.Time `json:"released"`
	} `json:"version"`
}

type JavaManifest struct {
	Files map[string]struct {
		Type       string `json:"type"`
		Executable bool   `json:"executable"`
		Target     string `json:"target"`
		Downloads  struct {
			LZMA struct {
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"lzma"`
			Raw struct {
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"raw"`
		} `json:"downloads"`
	} `json:"files"`
}

func (manifest JavaManifest) DownloadEntries(runtimeName string) (entries []network.DownloadEntry, symlinks map[string]string) {
	symlinks = make(map[string]string)
	dir := filepath.Join(env.JavaDir, runtimeName)
	for name, file := range manifest.Files {
		path := filepath.Join(dir, name)

		switch file.Type {
		case "link":
			if _, err := os.Lstat(path); err != nil {
				symlinks[path] = file.Target
			}
		case "file":
			data, err := os.ReadFile(path)
			if err == nil {
				sum := sha1.Sum(data)
				if hex.EncodeToString(sum[:]) == file.Downloads.Raw.Sha1 {
					continue
				}
			}
			var mode os.FileMode
			if file.Executable {
				mode = 0755
			}
			entries = append(entries, network.DownloadEntry{
				Sha1:     file.Downloads.Raw.Sha1,
				Path:     path,
				URL:      file.Downloads.Raw.URL,
				FileMode: mode,
			})
		}
	}
	return entries, symlinks
}

// A VersionMeta is metadata of the libraries, assets, and other data needed to start a Minecraft version.
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
	LoaderID    string `json:"-"`
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
	MainClass              string `json:"mainClass"`
	MinimumLauncherVersion int    `json:"minimumLauncherVersion"`
	ReleaseTime            string `json:"releaseTime"`
	Time                   string `json:"time"`
	Type                   string `json:"type"`
}

// Client creates a library from the client JAR download of versionMeta.
func (versionMeta VersionMeta) Client() Library {
	specifier, _ := NewLibrarySpecifier("com.mojang:minecraft:" + versionMeta.ID)
	return Library{
		Specifier: specifier,
		Artifact: Artifact{
			Path: fmt.Sprintf("com/mojang/minecraft/%s/%s.jar", versionMeta.ID, versionMeta.ID),
			Sha1: versionMeta.Downloads.Client.Sha1,
			Size: versionMeta.Downloads.Client.Size,
			URL:  versionMeta.Downloads.Client.URL,
		},
		ShouldInstall: true,
	}
}

// An Artifact represents a library JAR file that can be downloaded
type Artifact struct {
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

func (artifact Artifact) RuntimePath() string {
	return filepath.Join(env.LibrariesDir, artifact.Path)
}
func (artifact Artifact) IsDownloaded() bool {
	data, err := os.ReadFile(artifact.RuntimePath())
	if err != nil {
		return false
	}
	// if no checksum is present, still count the artifact as installed as long as the file exists
	if artifact.Sha1 == "" {
		return true
	}

	sum := sha1.Sum(data)
	return artifact.Sha1 == hex.EncodeToString(sum[:])
}
func (artifact Artifact) DownloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:  artifact.URL,
		Path: artifact.RuntimePath(),
		Sha1: artifact.Sha1,
	}
}

// A Library represents metadata of a game library and its artifact(s).
type Library struct {
	Artifact        Artifact
	Natives         []Library
	Specifier       LibrarySpecifier
	ShouldInstall   bool
	SkipOnClasspath bool
}

func (l *Library) UnmarshalJSON(b []byte) error {
	type library struct {
		Name      LibrarySpecifier `json:"name"`
		Downloads struct {
			Artifact    Artifact            `json:"artifact"`
			Classifiers map[string]Artifact `json:"classifiers"`
		} `json:"downloads"`
		Rules []struct {
			Action string `json:"action"`
			Os     struct {
				Name string `json:"name"`
			} `json:"os"`
		} `json:"rules,omitempty"`
		Natives map[string]string `json:"natives,omitempty"`

		// fabric

		URL  string `json:"url,omitempty"`
		Sha1 string `json:"sha1,omitempty"`
		Size int    `json:"size,omitempty"`
	}
	var data library
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	l.Specifier = data.Name
	if data.URL != "" {
		l.Artifact = Artifact{
			Path: data.Name.Path(),
			Sha1: data.Sha1,
			Size: data.Size,
			URL:  data.URL + "/" + data.Name.Path(),
		}
		l.ShouldInstall = true
	} else {
		l.Artifact = data.Downloads.Artifact
		var classifiers []string
		for os, native := range data.Natives {
			os = strings.ReplaceAll(os, "osx", "darwin")
			if os == runtime.GOOS {
				classifiers = append(classifiers, native)
			}
		}
		for _, classifier := range classifiers {
			artifact := data.Downloads.Classifiers[classifier]
			if artifact.URL == "" {
				continue
			}
			specifier := data.Name
			specifier.Classifier = classifier
			l.Natives = append(l.Natives, Library{
				Artifact:      artifact,
				Specifier:     specifier,
				ShouldInstall: true,
			})
		}
		if len(data.Rules) > 0 {
			rule := data.Rules[0]
			os := strings.ReplaceAll(rule.Os.Name, "osx", "darwin")
			if rule.Action == "allow" && (os == runtime.GOOS || os == "") {
				l.ShouldInstall = true
			}
		} else {
			l.ShouldInstall = true
		}
	}
	return nil
}

// An AssetIndex contains a map of asset objects and their names.
type AssetIndex struct {
	Objects map[string]struct {
		Hash string `json:"hash"`
		Size int    `json:"size"`
	} `json:"objects"`
}

// DownloadEntries returns a list of download entries to any undownloaded assets in the index.
func (index AssetIndex) DownloadEntries() (entries []network.DownloadEntry) {
	for _, object := range index.Objects {
		url, _ := url.JoinPath(MinecraftResourcesURL, object.Hash[:2], object.Hash)
		path := filepath.Join(env.AssetsDir, "objects", object.Hash[:2], object.Hash)
		data, err := os.ReadFile(path)
		if err == nil {
			sum := sha1.Sum(data)
			if object.Hash == hex.EncodeToString(sum[:]) {
				continue
			}
		}
		entries = append(entries, network.DownloadEntry{
			URL:  url,
			Path: path,
			Sha1: object.Hash,
		})
	}
	return entries
}

// FetchVersionManifest retrieves the Mojang version manifest which lists all game versions.
func FetchVersionManifest() (VersionManifest, error) {
	cache := network.JSONCache[VersionManifest]{
		Path: filepath.Join(env.CachesDir, "minecraft", "version_manifest.json"),
		URL:  VersionManifestURL,
	}

	var manifest VersionManifest

	if err := cache.FetchAndRead(&manifest); err != nil {
		if err := cache.Read(&manifest); err != nil {
			return VersionManifest{}, err
		}
	}
	return manifest, nil
}

// FetchVersionMeta retrieves the version metadata for a specified version from the version manifest.
//
// Besides normal version identifiers, "release" and "snapshot" are also accepted IDs.
func FetchVersionMeta(id string) (VersionMeta, error) {
	manifest, err := FetchVersionManifest()
	if err != nil {
		return VersionMeta{}, fmt.Errorf("retrieve version manifest: %w", err)
	}
	switch id {
	case "release":
		id = manifest.Latest.Release
	case "snapshot":
		id = manifest.Latest.Snapshot
	}
	for _, v := range manifest.Versions {
		if v.ID == id {
			cache := network.JSONCache[VersionMeta]{
				Path: filepath.Join(env.CachesDir, "minecraft", v.ID+".json"),
				URL:  v.URL,
			}
			download := true

			var versionMeta VersionMeta
			if err := cache.Read(&versionMeta); err == nil {
				sum, _ := cache.Sha1()
				if sum == v.Sha1 {
					download = false
				}
			}
			if download {
				if err := cache.FetchAndRead(&versionMeta); err != nil {
					return VersionMeta{}, err
				}
			}
			return versionMeta, nil
		}
	}
	return VersionMeta{}, fmt.Errorf("invalid version")
}

// MergeVersionMeta takes two instances of VersionMeta and merges w into v
func MergeVersionMeta(v, w VersionMeta) VersionMeta {
	v.Arguments.Jvm = w.Arguments.Jvm
	v.Arguments.Game = w.Arguments.Game

	m := make(map[string]int)
	for _, library := range w.Libraries {
		if !library.SkipOnClasspath {
			m[library.Specifier.Artifact]++
		}
	}

	libraries := w.Libraries

	for _, library := range v.Libraries {
		if m[library.Specifier.Artifact] < 1 {
			libraries = append(libraries, library)
		}
	}
	v.Libraries = libraries
	v.LoaderID = w.LoaderID
	if w.MainClass != "" {
		v.MainClass = w.MainClass
	}
	return v
}

// DownloadAssetIndex retrieves the asset index for the specified version.
func DownloadAssetIndex(versionMeta VersionMeta) (AssetIndex, error) {
	cache := network.JSONCache[AssetIndex]{
		Path: filepath.Join(env.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json"),
		URL:  versionMeta.AssetIndex.URL,
	}
	download := true
	var assetIndex AssetIndex
	if err := cache.Read(&assetIndex); err == nil {
		sum, _ := cache.Sha1()
		if sum == versionMeta.AssetIndex.Sha1 {
			download = false
		}
	}
	if download {
		if err := cache.FetchAndRead(&assetIndex); err != nil {
			return AssetIndex{}, err
		}
	}

	return assetIndex, nil
}

func FetchJavaManifestList() (JavaManifestList, error) {
	cache := network.JSONCache[JavaManifestList]{
		Path: filepath.Join(env.CachesDir, "minecraft", "java_all.json"),
		URL:  JavaRuntimesURL,
	}
	var list JavaManifestList
	if err := cache.Read(&list); err != nil {
		if err := cache.FetchAndRead(&list); err != nil {
			return JavaManifestList{}, err
		}
	}
	return list, nil
}

func FetchJavaManifest(name string) (JavaManifest, error) {
	list, err := FetchJavaManifestList()
	if err != nil {
		return JavaManifest{}, fmt.Errorf("retrieve java manifest list: %w", err)
	}

	os := strings.ReplaceAll(runtime.GOOS, "darwin", "mac-os")
	arch := strings.ReplaceAll(runtime.GOARCH, "386", "i386")

	if arch != "amd64" {
		os = os + "-" + arch
	}

	_, ok := list[os]
	if !ok {
		return JavaManifest{}, fmt.Errorf("system is unsupported")

	}
	_, ok = list[os][name]
	if !ok {
		return JavaManifest{}, fmt.Errorf("invalid name")
	}

	if len(list[os][name]) < 1 {
		return JavaManifest{}, fmt.Errorf("required version unavailable for this system")
	}

	ref := list[os][name][0].Manifest

	cache := network.JSONCache[JavaManifest]{
		Path: filepath.Join(env.CachesDir, "minecraft", "java-"+name+".json"),
		URL:  ref.URL,
	}

	download := true
	var manifest JavaManifest
	if err := cache.Read(&manifest); err == nil {
		sum, _ := cache.Sha1()
		if sum == ref.Sha1 {
			download = false
		}
	}

	if download {
		if err := cache.FetchAndRead(&manifest); err != nil {
			return JavaManifest{}, err
		}
	}
	return manifest, nil
}
