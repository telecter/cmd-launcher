package launcher

import (
	util "cmd-launcher/internal"
	"cmd-launcher/pkg/api"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func fetchLibraries(libraries []api.Library, rootDir string) ([]string, error) {
	var paths []string
	for _, library := range libraries {
		libraryPath := rootDir + "/libraries/" + library.Downloads.Artifact.Path

		err := util.DownloadFile(library.Downloads.Artifact.URL, libraryPath)
		if err != nil {
			return paths, fmt.Errorf("Error while downloading library %v: %v", library.Name, err)
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
		util.DownloadFile(fmt.Sprintf("https://resources.download.minecraft.net/%v/%v", beginOfHash, asset.Hash), fmt.Sprintf("%v/assets/objects/%v/%v", rootDir, beginOfHash, asset.Hash))
	}
	util.DownloadFile(meta.AssetIndex.URL, fmt.Sprintf("%v/assets/indexes/%v.json", rootDir, meta.AssetIndex.ID))
}

func Launch(version string) error {
	cwd, _ := os.Getwd()
	rootDir := cwd + "/minecraft"
	instanceDir := rootDir + "/instances/" + version

	err := os.MkdirAll(instanceDir, os.ModePerm)
	if err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("Could not create game directory: %v", err)
		}
	}

	meta, err := api.GetVersionMeta(version)
	if err != nil {
		return err
	}

	paths, err := fetchLibraries(meta.Libraries, rootDir)
	if err != nil {
		return err
	}

	err = util.DownloadFile(meta.Downloads.Client.URL, instanceDir+"/client.jar")
	if err != nil {
		return fmt.Errorf("Error while downloading game client: %v", err)
	}

	paths = append(paths, instanceDir+"/client.jar")

	fetchAssets(meta, rootDir)

	classPath := strings.Join(paths, ":")

	os.Chdir(instanceDir)
	cmd := exec.Command("java", "-XstartOnFirstThread", "-cp", classPath, meta.MainClass, "--accessToken", "abc", "--version", "", "--assetsDir", rootDir+"/assets", "--assetIndex", meta.AssetIndex.ID)
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
		fmt.Errorf("Error waiting for command: %v", err)
		return err
	}
	return nil
}
