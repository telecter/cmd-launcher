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

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/network/api"
)

type LaunchOptions struct {
	ModLoader       string
	QuickPlayServer string
}

func GetVersionDir(rootDir string, version string) string {
	return filepath.Join(rootDir, "versions", version)
}

func run(args []string) error {
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

func Launch(version string, options LaunchOptions, authData auth.MinecraftLoginData) error {
	if version == "latest" {
		latestVersion, err := api.GetLatestRelease()
		if err != nil {
			return err
		}
		version = latestVersion
	}
	versionDir := GetVersionDir(env.RootDir, version)

	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create game directory: %w", err)
	}

	var meta api.VersionMeta
	if data, err := os.ReadFile(filepath.Join(versionDir, version+".json")); err == nil {
		json.Unmarshal(data, &meta)
	} else {
		meta, err = api.GetVersionMeta(version)
		if err != nil {
			return err
		}
		json, _ := json.Marshal(meta)
		os.WriteFile(filepath.Join(versionDir, version+".json"), json, 0644)
	}

	libraries, err := installLibraries(meta.Libraries)
	if err != nil {
		return err
	}

	var loaderMeta api.FabricMeta
	if options.ModLoader == "fabric" {
		if data, err := os.ReadFile(filepath.Join(versionDir, "fabric.json")); err == nil {
			json.Unmarshal(data, &loaderMeta)
		} else {
			loaderMeta, err = api.GetLoaderMeta(version)
			if err != nil {
				return err
			}
			data, _ := json.Marshal(loaderMeta)
			os.WriteFile(filepath.Join(versionDir, "fabric.json"), data, 0644)
		}
		loaderLibraries, err := installLibraries(loaderMeta.Libraries)
		if err != nil {
			return err
		}
		libraries = append(libraries, loaderLibraries...)
	}

	if err = downloadAssets(meta); err != nil {
		return err
	}
	if err := downloadClient(meta, version); err != nil {
		return err
	}

	libraries = append(libraries, filepath.Join(versionDir, version+".jar"))

	jvmArgs := []string{"-cp", strings.Join(libraries, ":")}

	if runtime.GOOS == "darwin" {
		jvmArgs = append(jvmArgs, "-XstartOnFirstThread")
	}
	if options.ModLoader == "fabric" {
		jvmArgs = append(jvmArgs, loaderMeta.Arguments.Jvm...)
		jvmArgs = append(jvmArgs, loaderMeta.MainClass)
	} else {
		jvmArgs = append(jvmArgs, meta.MainClass)
	}

	gameArgs := []string{"--username", authData.Username, "--accessToken", authData.Token, "--gameDir", versionDir, "--assetsDir", filepath.Join(env.RootDir, "assets"), "--assetIndex", meta.AssetIndex.ID, "--version", version, "--versionType", meta.Type}
	if authData.UUID != "" {
		gameArgs = append(gameArgs, "--uuid", authData.UUID)
	}
	os.Chdir(versionDir)
	return run(append(jvmArgs, gameArgs...))
}
