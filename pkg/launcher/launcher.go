package launcher

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	util "github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/pkg/api"
)

type LaunchOptions struct {
	ModLoader string
}

func fetchLibraries(libraries []api.Library, rootDir string) ([]string, error) {
	var paths []string
	for _, library := range libraries {
		libraryPath := rootDir + "/libraries/" + library.Downloads.Artifact.Path

		err := util.DownloadFile(library.Downloads.Artifact.URL, libraryPath)
		if err != nil {
			return paths, fmt.Errorf("error while downloading library %s: %s", library.Name, err)
		}
		paths = append(paths, libraryPath)
	}
	return paths, nil
}

func fetchFabricLibraries(libraries []api.FabricLibrary, rootDir string) ([]string, error) {
	var paths []string
	for _, library := range libraries {
		libraryPath := rootDir + "/libraries/" + util.GetPathFromMaven(library.Name)
		err := util.DownloadFile(library.URL+util.GetPathFromMaven(library.Name), libraryPath)
		if err != nil {
			return paths, fmt.Errorf("error while downloading Fabric/Quilt library %s: %s", library.Name, err)
		}
		paths = append(paths, libraryPath)
	}
	return paths, nil
}

func fetchAssets(meta api.VersionMeta, rootDir string) {
	index := api.AssetIndex{}
	util.GetJSON(meta.AssetIndex.URL, &index)
	for _, asset := range index.Objects {
		beginOfHash := asset.Hash[:2]
		util.DownloadFile(fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", beginOfHash, asset.Hash), fmt.Sprintf("%v/assets/objects/%v/%v", rootDir, beginOfHash, asset.Hash))
	}
	util.DownloadFile(meta.AssetIndex.URL, fmt.Sprintf("%v/assets/indexes/%v.json", rootDir, meta.AssetIndex.ID))
}

func Launch(version string, rootDir string, options LaunchOptions, authData api.AuthData) error {
	instanceDir := rootDir + "/instances/" + version

	err := os.MkdirAll(instanceDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create game directory: %s", err)
	}

	meta, err := api.GetVersionMeta(version)
	if err != nil {
		return err
	}

	paths, err := fetchLibraries(meta.Libraries, rootDir)
	if err != nil {
		return err
	}

	var loaderMeta api.FabricMeta
	if options.ModLoader != "" {
		var url string
		if options.ModLoader == "fabric" {
			url = api.FabricURLPrefix
		} else if options.ModLoader == "quilt" {
			url = api.QuiltURLPrefix
		} else {
			return fmt.Errorf("invalid mod loader")
		}
		loaderMeta, err = api.GetLoaderMeta(url, version)
		if err != nil {
			return err
		}
		fabricPaths, err := fetchFabricLibraries(loaderMeta.Libraries, rootDir)
		paths = append(paths, fabricPaths...)
		if err != nil {
			return err
		}
	}

	err = util.DownloadFile(meta.Downloads.Client.URL, instanceDir+"/client.jar")
	if err != nil {
		return fmt.Errorf("error while downloading game client: %s", err)
	}

	paths = append(paths, instanceDir+"/client.jar")

	fetchAssets(meta, rootDir)

	classPath := strings.Join(paths, ":")

	os.Chdir(instanceDir)

	jvmArgs := []string{"-cp", classPath}
	if runtime.GOOS == "darwin" {
		jvmArgs = append([]string{"-XstartOnFirstThread"}, jvmArgs...)
	}
	if options.ModLoader == "fabric" || options.ModLoader == "quilt" {
		jvmArgs = append(jvmArgs, "-DFabricMcEmu= net.minecraft.client.main.Main")
		jvmArgs = append(jvmArgs, loaderMeta.MainClass)
	} else {
		jvmArgs = append(jvmArgs, meta.MainClass)
	}

	gameArgs := []string{"--accessToken", authData.Token, "--version", "", "--assetsDir", rootDir + "/assets", "--assetIndex", meta.AssetIndex.ID, "--username", authData.Username, "--uuid", authData.UUID}

	args := append(jvmArgs, gameArgs...)

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

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for command: %v", err)
	}
	return nil
}
