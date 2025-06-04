// Package launcher provides the necessary functions to start the game.
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
	"golang.org/x/mod/semver"
)

// Loader represents a game mod loader.
type Loader string

const (
	LoaderVanilla Loader = "vanilla"
	LoaderFabric  Loader = "fabric"
	LoaderQuilt   Loader = "quilt"
)

func (loader Loader) String() string {
	return string(loader)
}

type EnvOptions struct {
	Session            auth.Session
	Config             InstanceConfig
	QuickPlayServer    string
	Demo               bool
	DisableMultiplayer bool
	DisableChat        bool
}

// A launchEnvironment represents the data needed to start the game.
type launchEnvironment struct {
	gameDir   string
	javaPath  string
	mainClass string
	classpath []string
	javaArgs  []string
	gameArgs  []string
}

// An EventWatcher is a controller that handles game preparation events.
type EventWatcher interface {
	Handle(event any)
}

// LibrariesResolvedEvent is created when all game libraries have been identified and filtered.
type LibrariesResolvedEvent struct {
	Libraries int
}

// AssetsResolvedEvent is called when all game assets have been identified and filtered.
type AssetsResolvedEvent struct {
	Assets int
}

// DownloadingEvent is called when a download has progressed.
type DownloadingEvent struct {
	Completed int
	Total     int
}

// A Runner is a controller which manages the starting of the game.
type Runner interface {
	Run(cmd *exec.Cmd) error
}

// An ConsoleRunner is an implementation of Runner which logs game output to the console.
type ConsoleRunner struct{}

func (ConsoleRunner) Run(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// filterLibraries transforms libraries into installed and required libraries.
//
// It also provides fixes and swaps out some libraries to ensure compatability.
func filterLibraries(mojang []meta.Library, external ...meta.Library) (installed []meta.Library, required []meta.Library) {
	m := make(map[string]int)
	for _, library := range external {
		m[library.Name.Group+":"+library.Name.Artifact+":"+library.Name.Classifier]++
	}
	for i, library := range mojang {
		if m[library.Name.Group+":"+library.Name.Artifact+":"+library.Name.Classifier] > 0 {
			mojang = slices.Delete(mojang, i, i+1)
		}
	}
	for _, library := range append(mojang, external...) {
		if library.ShouldInstall() {
			if library.Name.Group == "org.lwjgl" {
				fmt.Println(library.Name.Version)
				if library.Name.Classifier == "natives-linux" && runtime.GOARCH == "arm64" && library.Name.Artifact == "lwjgl-jemalloc" && semver.Compare("v3.3.2", "v"+library.Name.Version) == 1 {
					specifier := library.Name
					specifier.Version = "3.2.2"
					library = meta.Library{
						Name: specifier,
						Artifact: meta.Artifact{
							URL:  "https://raw.githubusercontent.com/theofficialgman/lwjgl3-binaries-arm64/refs/heads/lwjgl-3.2.2/lwjgl-jemalloc-patched-natives-linux-arm64.jar",
							Sha1: "762d7d80c9cdf3a3f3fc80c8a5f86612255edfe0",
							Path: specifier.Path(),
						},
					}
				} else {
					runtimeOs := runtime.GOOS
					if runtime.GOARCH != "amd64" {
						runtimeOs += "-" + runtime.GOARCH
					}
					if library.Name.Classifier != "" {
						library.Name.Classifier = "natives-" + runtimeOs
					}
					library = meta.GetMavenLibrary(library.Name)
				}
			}
			if library.IsInstalled() {
				installed = append(installed, library)
			} else {
				required = append(required, library)
			}
		}
	}
	return installed, required
}

// Launch starts a launchEnvironment with the specified runner.
//
// The Java executable is checked and the classpath and command arguments are finalized.
func Launch(launchEnv launchEnvironment, runner Runner) error {
	info, err := os.Stat(launchEnv.javaPath)
	if err != nil {
		return fmt.Errorf("java executable does not exist")
	}
	if info.Mode()&0111 == 0 || info.IsDir() {
		return fmt.Errorf("java binary is not executable")
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}
	javaArgs := append(launchEnv.javaArgs, "-cp", strings.Join(launchEnv.classpath, separator), launchEnv.mainClass)
	cmd := exec.Command(launchEnv.javaPath, append(javaArgs, launchEnv.gameArgs...)...)
	os.Chdir(launchEnv.gameDir)

	return runner.Run(cmd)
}

// Prepare prepares the specified instance to be launched, returning a launchEnvironment, with the provided options and sends events to watcher.
func Prepare(inst Instance, options EnvOptions, watcher EventWatcher) (launchEnvironment, error) {
	launchEnv := launchEnvironment{
		javaPath: options.Config.Java,
		gameDir:  inst.Dir,
	}

	versionMeta, err := meta.GetVersionMeta(inst.GameVersion)
	if err != nil {
		return launchEnvironment{}, fmt.Errorf("fetch version metadata: %w", err)
	}

	launchEnv.mainClass = versionMeta.MainClass

	var loaderLibraries []meta.Library
	if inst.Loader == LoaderFabric || inst.Loader == LoaderQuilt {
		var fabricLoader meta.FabricLoader
		switch inst.Loader {
		case LoaderFabric:
			fabricLoader = meta.FabricLoaderStandard
		case LoaderQuilt:
			fabricLoader = meta.FabricLoaderQuilt
		}
		fabricMeta, err := meta.GetFabricMeta(versionMeta.ID, inst.LoaderVersion, fabricLoader)
		if err != nil {
			return launchEnvironment{}, err
		}
		loaderLibraries = fabricMeta.Libraries
		launchEnv.javaArgs = append(launchEnv.javaArgs, fabricMeta.Arguments.Jvm...)
		launchEnv.mainClass = fabricMeta.MainClass
	}
	installedLibs, requiredLibs := filterLibraries(append(versionMeta.Libraries, versionMeta.Client()), loaderLibraries...)
	watcher.Handle(LibrariesResolvedEvent{
		Libraries: len(installedLibs) + len(requiredLibs),
	})

	var downloads []network.DownloadEntry

	for _, library := range requiredLibs {
		downloads = append(downloads, library.DownloadEntry())
	}
	assetIndex, err := meta.DownloadAssetIndex(versionMeta)
	if err != nil {
		return launchEnvironment{}, fmt.Errorf("fetch asset index: %w", err)
	}

	for _, object := range assetIndex.Objects {
		if !object.IsDownloaded() {
			downloads = append(downloads, object.DownloadEntry())
		}
	}

	if len(downloads) > 0 {
		results := network.StartDownloadEntries(downloads)
		i := 0
		for err := range results {
			if err != nil {
				return launchEnvironment{}, fmt.Errorf("download files: %w", err)
			}
			watcher.Handle(DownloadingEvent{
				Completed: i,
				Total:     len(downloads),
			})
			i++
		}
	}

	if runtime.GOOS == "darwin" {
		launchEnv.javaArgs = append(launchEnv.javaArgs, "-XstartOnFirstThread")
	}
	if options.Config.MinMemory != 0 {
		launchEnv.javaArgs = append(launchEnv.javaArgs, fmt.Sprintf("-Xms%dm", options.Config.MinMemory))
	}
	if options.Config.MaxMemory != 0 {
		launchEnv.javaArgs = append(launchEnv.javaArgs, fmt.Sprintf("-Xmx%dm", options.Config.MaxMemory))
	}
	launchEnv.gameArgs = []string{
		"--username", options.Session.Username,
		"--accessToken", options.Session.AccessToken,
		"--userType", "msa",
		"--gameDir", inst.Dir,
		"--assetsDir", env.AssetsDir,
		"--assetIndex", versionMeta.AssetIndex.ID,
		"--version", versionMeta.ID,
		"--versionType", versionMeta.Type,
		"--width", strconv.Itoa(options.Config.WindowResolution.Width),
		"--height", strconv.Itoa(options.Config.WindowResolution.Height),
	}
	if options.QuickPlayServer != "" {
		launchEnv.gameArgs = append(launchEnv.gameArgs, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.Session.UUID != "" {
		launchEnv.gameArgs = append(launchEnv.gameArgs, "--uuid", options.Session.UUID)
	}
	if options.Demo {
		launchEnv.gameArgs = append(launchEnv.gameArgs, "--demo")
	}
	if options.DisableChat {
		launchEnv.gameArgs = append(launchEnv.gameArgs, "--disableChat")
	}
	if options.DisableMultiplayer {
		launchEnv.gameArgs = append(launchEnv.gameArgs, "--disableMultiplayer")
	}

	for _, library := range append(installedLibs, requiredLibs...) {
		launchEnv.classpath = append(launchEnv.classpath, library.RuntimePath())
	}
	return launchEnv, nil
}
