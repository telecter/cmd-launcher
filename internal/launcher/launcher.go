package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/meta"
)

const (
	LoaderVanilla string = "vanilla"
	LoaderFabric  string = "fabric"
	LoaderQuilt   string = "quilt"
)

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
	instance, err := GetInstance(instanceId)
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

	versionMeta, err := meta.GetVersionMeta(instance.GameVersion)
	if err != nil {
		return err
	}

	var javaArgs []string
	mainClass := versionMeta.MainClass

	libraries := versionMeta.Libraries
	if instance.Loader == LoaderFabric {
		for i, library := range libraries {
			if strings.Contains(library.Name, "org.ow2.asm:asm:") {
				libraries = slices.Delete(libraries, i, i+1)
				break
			}
		}
	}

	if instance.Loader == LoaderFabric || instance.Loader == LoaderQuilt {
		var fabricMeta meta.FabricMeta
		if instance.Loader == LoaderFabric {
			fabricMeta, err = meta.GetFabricMeta(versionMeta.ID)
			if err != nil {
				return err
			}
		} else if instance.Loader == LoaderQuilt {
			fabricMeta, err = meta.GetQuiltMeta(versionMeta.ID)
			if err != nil {
				return err
			}
		}
		libraries = append(libraries, fabricMeta.Libraries...)
		javaArgs = append(javaArgs, fabricMeta.Arguments.Jvm...)
		mainClass = fabricMeta.MainClass
	}
	installed, required := filterLibraries(libraries)
	if err := installLibraries(required); err != nil {
		return err
	}

	libraryPaths := getRuntimeLibraryPaths(append(installed, required...))

	assetIndex, err := downloadAssetIndex(versionMeta)
	if err != nil {
		return fmt.Errorf("fetch asset index: %w", err)
	}

	requiredAssets := getRequiredAssets(assetIndex)
	if err := downloadAssets(requiredAssets); err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		javaArgs = append(javaArgs, "-XstartOnFirstThread")
	}
	if instance.Config.MinMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xms%dm", instance.Config.MinMemory))
	}
	if instance.Config.MaxMemory != 0 {
		javaArgs = append(javaArgs, fmt.Sprintf("-Xmx%dm", instance.Config.MaxMemory))
	}

	gameArgs := []string{
		"--username", options.LoginData.Username,
		"--accessToken", options.LoginData.Token,
		"--userType", "msa",
		"--gameDir", instance.Dir,
		"--assetsDir", internal.AssetsDir,
		"--assetIndex", versionMeta.AssetIndex.ID,
		"--version", versionMeta.ID,
		"--versionType", versionMeta.Type,
		"--width", strconv.Itoa(instance.Config.WindowResolution.Width),
		"--height", strconv.Itoa(instance.Config.WindowResolution.Height),
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
	os.Chdir(instance.Dir)
	return run(runOptions{
		javaPath:  instance.Config.JavaExecutablePath,
		mainClass: mainClass,
		classpath: libraryPaths,
		javaArgs:  javaArgs,
		gameArgs:  gameArgs,
	})
}
