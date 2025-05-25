package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type Loader string

const (
	LoaderVanilla Loader = "vanilla"
	LoaderFabric  Loader = "fabric"
	LoaderQuilt   Loader = "quilt"
)

func (loader Loader) String() string {
	return string(loader)
}

type LaunchOptions struct {
	QuickPlayServer    string
	LoginData          auth.MinecraftLoginData
	OfflineMode        bool
	Demo               bool
	DisableMultiplayer bool
	DisableChat        bool
}
type runOptions struct {
	javaPath  string
	mainClass string
	classpath []string
	javaArgs  []string
	gameArgs  []string
}

func run(options runOptions) error {
	info, err := os.Stat(options.javaPath)
	if err != nil {
		return fmt.Errorf("check Java executable: %w", err)
	}
	if info.Mode()&0111 == 0 || info.IsDir() {
		return fmt.Errorf("check Java executable: file is not an executable")
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}
	javaArgs := append(options.javaArgs, "-cp", strings.Join(options.classpath, separator), options.mainClass)

	cmd := exec.Command(options.javaPath, append(javaArgs, options.gameArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Launch(instanceId string, options LaunchOptions) error {
	inst, err := GetInstance(instanceId)
	if err != nil {
		return err
	}
	if !options.OfflineMode {
		loginData, err := auth.LoginWithMicrosoft()
		if err != nil {
			return fmt.Errorf("login with Microsoft: %w", err)
		}
		options.LoginData = loginData
	}

	versionMeta, err := meta.GetVersionMeta(inst.GameVersion)
	if err != nil {
		return err
	}

	var javaArgs []string
	mainClass := versionMeta.MainClass

	libraries := fixLibraries(versionMeta.Libraries, inst.Loader)

	if inst.Loader == LoaderFabric || inst.Loader == LoaderQuilt {
		var fabricLoader meta.FabricLoader
		switch inst.Loader {
		case LoaderFabric:
			fabricLoader = meta.FabricLoaderStandard
		case LoaderQuilt:
			fabricLoader = meta.FabricLoaderQuilt
		}
		var fabricMeta meta.FabricMeta
		fabricVersions, err := meta.GetFabricVersions(versionMeta.ID, fabricLoader)
		if err != nil {
			return err
		}
		fabricMeta, err = meta.GetFabricMeta(versionMeta.ID, fabricVersions[0].Loader.Version, fabricLoader)
		if err != nil {
			return err
		}
		libraries = append(libraries, fabricMeta.Libraries...)
		javaArgs = append(javaArgs, fabricMeta.Arguments.Jvm...)
		mainClass = fabricMeta.MainClass
	}
	installedLibs, requiredLibs := filterLibraries(append(libraries, getClientLibrary(versionMeta)))

	if len(requiredLibs) > 0 {
		bar := progressbar.Default(int64(len(requiredLibs)), "Installing libraries")
		for _, library := range requiredLibs {
			if err := library.Install(); err != nil {
				return fmt.Errorf("download library '%s': %w", library.Name, err)
			}
			bar.Add(1)
		}
	}

	assetIndex, err := downloadAssetIndex(versionMeta)
	if err != nil {
		return fmt.Errorf("fetch asset index: %w", err)
	}

	requiredAssets := filterAssets(assetIndex)

	if len(requiredAssets.Objects) > 0 {
		bar := progressbar.Default(int64(len(requiredAssets.Objects)), "Downloading assets")
		for _, asset := range requiredAssets.Objects {
			if err := downloadAsset(asset); err != nil {
				return err
			}
			bar.Add(1)
		}
	}

	if runtime.GOOS == "darwin" {
		javaArgs = append(javaArgs, "-XstartOnFirstThread")
	}
	if inst.Config.MinMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xms%dm", inst.Config.MinMemory))
	}
	if inst.Config.MaxMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xmx%dm", inst.Config.MaxMemory))
	}

	gameArgs := []string{
		"--username", options.LoginData.Username,
		"--accessToken", options.LoginData.Token,
		"--userType", "msa",
		"--gameDir", inst.Dir,
		"--assetsDir", internal.AssetsDir,
		"--assetIndex", versionMeta.AssetIndex.ID,
		"--version", versionMeta.ID,
		"--versionType", versionMeta.Type,
		"--width", strconv.Itoa(inst.Config.WindowResolution.Width),
		"--height", strconv.Itoa(inst.Config.WindowResolution.Height),
	}
	if options.QuickPlayServer != "" {
		gameArgs = append(gameArgs, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.LoginData.UUID != "" {
		gameArgs = append(gameArgs, "--uuid", options.LoginData.UUID)
	}
	if options.Demo {
		gameArgs = append(gameArgs, "--demo")
	}
	if options.DisableChat {
		gameArgs = append(gameArgs, "--disableChat")
	}
	if options.DisableMultiplayer {
		gameArgs = append(gameArgs, "--disableMultiplayer")
	}

	var classpath []string
	for _, library := range append(installedLibs, requiredLibs...) {
		classpath = append(classpath, library.RuntimePath())
	}
	os.Chdir(inst.Dir)
	return run(runOptions{
		javaPath:  inst.Config.JavaExecutablePath,
		mainClass: mainClass,
		classpath: classpath,
		javaArgs:  javaArgs,
		gameArgs:  gameArgs,
	})
}
