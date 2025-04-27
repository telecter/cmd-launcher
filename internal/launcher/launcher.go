package launcher

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type LaunchOptions struct {
	QuickPlayServer string
	LoginData       auth.MinecraftLoginData
	OfflineMode     bool
}

func run(javaPath string, args []string) error {
	cmd := exec.Command(javaPath, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go func() {
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	return cmd.Run()
}

func Launch(instanceId string, options LaunchOptions) error {
	instance, err := GetInstance(instanceId)
	if err != nil {
		return err
	}
	if !options.OfflineMode {
		loginData, err := auth.LoginWithMicrosoft()
		if err != nil {
			return fmt.Errorf("error logging in with Microsoft: %w", err)
		}
		options.LoginData = loginData
	}

	versionMeta, err := meta.GetVersionMeta(instance.GameVersion)
	if err != nil {
		return err
	}

	var javaArgs []string
	mainClass := versionMeta.MainClass

	allLibraries := append(versionMeta.Libraries, meta.Library{
		Name: "com.mojang:minecraft:" + versionMeta.ID,
		Downloads: struct {
			Artifact meta.Artifact "json:\"artifact\""
		}{
			Artifact: meta.Artifact{
				Path: fmt.Sprintf("com/mojang/minecraft/%s/%s.jar", versionMeta.ID, versionMeta.ID),
				Sha1: versionMeta.Downloads.Client.Sha1,
				Size: versionMeta.Downloads.Client.Size,
				URL:  versionMeta.Downloads.Client.URL,
			},
		},
	})

	if instance.ModLoader == "fabric" {
		fabricMeta, err := meta.GetFabricMeta(versionMeta.ID)
		if err != nil {
			return err
		}
		allLibraries = append(allLibraries, fabricMeta.Libraries...)
		javaArgs = append(javaArgs, fabricMeta.Arguments.Jvm...)
		mainClass = fabricMeta.MainClass
	}
	installed, required := filterLibraries(allLibraries)
	if err := installLibraries(required); err != nil {
		return err
	}

	libraries := append(installed, required...)
	libraryPaths := getRuntimeLibraryPaths(libraries)

	// TEMPORARY FIX: Duplicate ASM classes
	if instance.ModLoader == "fabric" {
		for i, libraryPath := range libraryPaths {
			if strings.Contains(libraryPath, "asm-9.6.jar") {
				libraryPaths = slices.Delete(libraryPaths, i, i+1)
				break
			}
		}
	}

	assetIndex, err := downloadAssetIndex(versionMeta)
	if err != nil {
		return fmt.Errorf("failed to get asset index: %w", err)
	}

	requiredAssetIndex := getRequiredAssets(assetIndex)
	if err := downloadAssets(requiredAssetIndex); err != nil {
		return err
	}

	javaArgs = append(javaArgs, "-cp")
	javaArgs = append(javaArgs, strings.Join(libraryPaths, ":"))

	if runtime.GOOS == "darwin" {
		javaArgs = append(javaArgs, "-XstartOnFirstThread")
	}
	if instance.Config.MinMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xms%dm", instance.Config.MinMemory))
	}
	if instance.Config.MaxMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xmx%dm", instance.Config.MaxMemory))
	}
	javaArgs = append(javaArgs, mainClass)

	gameArgs := []string{
		"--username", options.LoginData.Username,
		"--accessToken", options.LoginData.Token,
		"--userType", "msa",
		"--gameDir", instance.Dir,
		"--assetsDir", env.AssetsDir,
		"--assetIndex", versionMeta.AssetIndex.ID,
		"--version", versionMeta.ID,
		"--versionType", versionMeta.Type,
		"--width", strconv.Itoa(instance.Config.WindowResolution[0]),
		"--height", strconv.Itoa(instance.Config.WindowResolution[1]),
	}
	if options.QuickPlayServer != "" {
		gameArgs = append(gameArgs, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.LoginData.UUID != "" {
		gameArgs = append(gameArgs, "--uuid", options.LoginData.UUID)
	}
	os.Chdir(instance.Dir)
	return run(instance.Config.JavaExecutablePath, append(javaArgs, gameArgs...))
}
