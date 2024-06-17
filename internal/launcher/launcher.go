package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	util "github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/api"
)

type LaunchOptions struct {
	ModLoader       string
	QuickPlayServer string
}

func getLibraries(libraries []api.Library, rootDir string) ([]string, error) {
	var paths []string
	for _, library := range libraries {
		if len(library.Rules) > 0 {
			os := library.Rules[0].Os.Name
			os = strings.ReplaceAll(os, "osx", "darwin")
			if os != "" && os != runtime.GOOS {
				continue
			}
		}
		var url string
		var file string
		if library.URL != "" {
			url = library.URL + util.GetPathFromMaven(library.Name)
			file = filepath.Join(rootDir, "libraries", util.GetPathFromMaven(library.Name))
		} else {
			url = library.Downloads.Artifact.URL
			file = filepath.Join(rootDir, "libraries", library.Downloads.Artifact.Path)
		}
		err := util.DownloadFile(url, file)
		if err != nil {
			return paths, fmt.Errorf("error while downloading %s: %s", library.Name, err)
		}
		paths = append(paths, file)
	}
	return paths, nil
}

func getAssets(meta api.VersionMeta, rootDir string) error {
	index := api.AssetIndex{}
	indexPath := filepath.Join(rootDir, "assets", "indexes", meta.AssetIndex.ID+".json")

	if data, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(data, &index)
	} else {
		if err := util.GetJSON(meta.AssetIndex.URL, &index); err != nil {
			return fmt.Errorf("error while getting asset index: %s", err)
		}
	}
	for _, asset := range index.Objects {
		prefix := asset.Hash[:2]
		util.DownloadFile(fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", prefix, asset.Hash), filepath.Join(rootDir, "assets", "objects", prefix, asset.Hash))
	}

	util.DownloadFile(meta.AssetIndex.URL, indexPath)
	return nil
}

func runJava(args []string) error {
	cmd := exec.Command("java", args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	cmd.Start()

	go func() {
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	return cmd.Wait()
}

func Launch(version string, rootDir string, options LaunchOptions, authData api.AuthData) error {
	versionDir := util.GetInstanceDir(rootDir, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("error creating game directory: %s", err)
	}

	var meta api.VersionMeta
	if data, err := os.ReadFile(filepath.Join(versionDir, version+".json")); err == nil {
		json.Unmarshal(data, &meta)
	} else {
		meta, err = api.GetVersionMeta(version)
		if err != nil {
			return err
		}
		version = meta.ID
	}

	libraries, err := getLibraries(meta.Libraries, rootDir)
	if err != nil {
		return fmt.Errorf("error downloading libraries: %s", err)
	}

	var loaderMeta api.FabricMeta
	if options.ModLoader != "" {
		var url string
		switch options.ModLoader {
		case "fabric":
			url = api.FabricURLPrefix
		case "quilt":
			url = api.QuiltURLPrefix
		default:
			return fmt.Errorf("invalid mod loader")
		}
		if data, err := os.ReadFile(filepath.Join(versionDir, options.ModLoader+".json")); err == nil {
			json.Unmarshal(data, &loaderMeta)
		} else {
			loaderMeta, err = api.GetLoaderMeta(url, version)
			if err != nil {
				return err
			}
			data, _ := json.Marshal(loaderMeta)
			os.WriteFile(filepath.Join(versionDir, options.ModLoader+".json"), data, 0644)
		}
		loaderLibraries, err := getLibraries(loaderMeta.Libraries, rootDir)
		if err != nil {
			return fmt.Errorf("error downloading loader libraries: %s", err)
		}
		libraries = append(libraries, loaderLibraries...)
	}

	if err = getAssets(meta, rootDir); err != nil {
		return fmt.Errorf("error downloading assets: %s", err)
	}

	if err := util.DownloadFile(meta.Downloads.Client.URL, filepath.Join(versionDir, version+".jar")); err != nil {
		return fmt.Errorf("error downloading client: %s", err)
	}
	libraries = append(libraries, filepath.Join(versionDir, version+".jar"))

	jvmArgs := []string{"-cp", strings.Join(libraries, ":")}

	if runtime.GOOS == "darwin" {
		jvmArgs = append(jvmArgs, "-XstartOnFirstThread")
	}
	if options.ModLoader != "" {
		jvmArgs = append(jvmArgs, loaderMeta.Arguments.Jvm...)
		jvmArgs = append(jvmArgs, loaderMeta.MainClass)
		fmt.Println(loaderMeta.MainClass)
	} else {
		jvmArgs = append(jvmArgs, meta.MainClass)
	}

	gameArgs := []string{"--username", authData.Username, "--accessToken", authData.Token, "--gameDir", versionDir, "--assetsDir", filepath.Join(rootDir, "assets"), "--assetIndex", meta.AssetIndex.ID, "--version", ""}
	if authData.UUID != "" {
		gameArgs = append(gameArgs, "--uuid", authData.UUID)
	}

	os.Chdir(versionDir)
	fmt.Println(jvmArgs, gameArgs)
	return runJava(append(jvmArgs, gameArgs...))
}
