// Package launcher provides the necessary functions to start the game.
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

// Loader represents a game mod loader.
type Loader string

const (
	LoaderVanilla  Loader = "vanilla"
	LoaderFabric   Loader = "fabric"
	LoaderQuilt    Loader = "quilt"
	LoaderNeoForge Loader = "neoforge"
	LoaderForge    Loader = "forge"
)

func (loader Loader) String() string {
	return string(loader)
}

// EnvOptions represents configuration options when preparing an instance to be launched.
type EnvOptions struct {
	Session            auth.Session
	Config             InstanceConfig
	QuickPlayServer    string
	Demo               bool
	DisableMultiplayer bool
	DisableChat        bool

	skipAssets    bool
	skipLibraries bool
}

// An EventWatcher is a controller that can handle multiple types of events.
type EventWatcher interface {
	Handle(event any)
}

// MetadataResolvedEvent is called when all metadata has been retrieved
type MetadataResolvedEvent struct{}

// LibrariesResolvedEvent is called when all game libraries have been identified and filtered.
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

// A LaunchEnvironment represents the data needed to start the game.
type LaunchEnvironment struct {
	GameDir   string
	JavaPath  string
	MainClass string
	Classpath []string
	JavaArgs  []string
	GameArgs  []string
}

// Launch starts a LaunchEnvironment with the specified runner.
//
// The Java executable is checked and the classpath and command arguments are finalized.
func Launch(launchEnv LaunchEnvironment, runner Runner) error {
	info, err := os.Stat(launchEnv.JavaPath)
	if err != nil {
		return fmt.Errorf("java executable does not exist")
	}
	if info.Mode()&0111 == 0 || info.IsDir() {
		return fmt.Errorf("java binary is not executable")
	}

	javaArgs := append(launchEnv.JavaArgs, "-cp", strings.Join(launchEnv.Classpath, string(os.PathListSeparator)), launchEnv.MainClass)
	cmd := exec.Command(launchEnv.JavaPath, append(javaArgs, launchEnv.GameArgs...)...)
	cmd.Dir = launchEnv.GameDir
	return runner.Run(cmd)
}

// Prepare prepares the instance to be launched, returning a LaunchEnvironment, with the provided options and sends events to watcher.
func Prepare(inst Instance, options EnvOptions, watcher EventWatcher) (LaunchEnvironment, error) {
	var downloads []network.DownloadEntry

	version, err := fetchVersion(inst.Loader, inst.GameVersion, inst.LoaderVersion)
	if err != nil {
		return LaunchEnvironment{}, fmt.Errorf("retrieve loader metadata: %w", err)
	}
	version.Libraries = append(version.Libraries, version.Client())

	launchEnv := LaunchEnvironment{
		GameDir:   inst.Dir(),
		JavaPath:  options.Config.Java,
		MainClass: version.MainClass,
	}
	watcher.Handle(MetadataResolvedEvent{})

	// Filter libraries, and add necessary artifact download entries
	installedLibs, requiredLibs := filterLibraries(version.Libraries)
	if !options.skipLibraries {
		for _, library := range requiredLibs {
			downloads = append(downloads, library.Artifact.DownloadEntry())
		}
	}
	watcher.Handle(LibrariesResolvedEvent{
		Libraries: len(installedLibs) + len(requiredLibs),
	})

	// Download asset index and add all necessary asset download entries
	assetIndex, err := meta.DownloadAssetIndex(version)
	if err != nil {
		return LaunchEnvironment{}, fmt.Errorf("retrieve asset index: %w", err)
	}
	if !options.skipAssets {
		downloads = append(downloads, assetIndex.DownloadEntries()...)
	}
	watcher.Handle(AssetsResolvedEvent{Assets: len(assetIndex.Objects)})

	if launchEnv.JavaPath == "" {
		manifest, err := meta.FetchJavaManifest(version.JavaVersion.Component)
		if err != nil {
			return LaunchEnvironment{}, fmt.Errorf("fetch java manifest: %w", err)
		}
		entries, symlinks := manifest.DownloadEntries(version.JavaVersion.Component)
		downloads = append(downloads, entries...)
		for link, target := range symlinks {
			if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
				panic(err)
			}

			if err := os.Symlink(target, link); err != nil {
				return LaunchEnvironment{}, fmt.Errorf("create symlink %q: %w", link, err)
			}
		}
		launchEnv.JavaPath = filepath.Join(env.JavaDir, version.JavaVersion.Component, "bin", "java")
	}

	if err := download(downloads, watcher); err != nil {
		return LaunchEnvironment{}, fmt.Errorf("download files: %w", err)
	}

	var processors []meta.ForgeProcessor
	if inst.Loader == LoaderForge {
		processors, err = meta.Forge.FetchPostProcessors(version.ID, version.LoaderID)
		if err != nil {
			return LaunchEnvironment{}, fmt.Errorf("fetch forge post processors: %w", err)
		}
	} else if inst.Loader == LoaderNeoForge {
		processors, err = meta.Neoforge.FetchPostProcessors(version.ID, version.LoaderID)
		if err != nil {
			return LaunchEnvironment{}, fmt.Errorf("fetch neoforge post processors: %w", err)
		}
	}
	if err := postProcess(launchEnv, processors); err != nil {
		return LaunchEnvironment{}, fmt.Errorf("run forge post processors: %w", err)
	}

	launchEnv.JavaArgs, launchEnv.GameArgs = createArgs(launchEnv, version, options)

	for _, library := range append(installedLibs, requiredLibs...) {
		if library.SkipOnClasspath {
			continue
		}
		launchEnv.Classpath = append(launchEnv.Classpath, library.Artifact.RuntimePath())
	}

	return launchEnv, nil
}

// download takes a list of download entries and executes them, reporting download events to watcher.
func download(entries []network.DownloadEntry, watcher EventWatcher) error {
	if len(entries) > 0 {
		results := network.StartDownloadEntries(entries)
		i := 0
		for err := range results {
			if err != nil {
				return err
			}
			watcher.Handle(DownloadingEvent{
				Completed: i,
				Total:     len(entries),
			})
			i++
		}
	}
	return nil
}

// createArgs takes data from a launch environment, version metadata, and environment options to
// create a set of game and Java arguments to pass when starting the game.
func createArgs(launchEnv LaunchEnvironment, version meta.VersionMeta, options EnvOptions) (java, game []string) {
	game = []string{
		"--username", options.Session.Username,
		"--accessToken", options.Session.AccessToken,
		"--userType", "msa",
		"--gameDir", launchEnv.GameDir,
		"--assetsDir", env.AssetsDir,
		"--assetIndex", version.AssetIndex.ID,
		"--version", version.ID,
		"--versionType", version.Type,
		"--width", strconv.Itoa(options.Config.WindowResolution.Width),
		"--height", strconv.Itoa(options.Config.WindowResolution.Height),
	}
	if options.QuickPlayServer != "" {
		game = append(game, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.Session.UUID != "" {
		game = append(game, "--uuid", options.Session.UUID)
	}
	if options.Demo {
		game = append(game, "--demo")
	}
	if options.DisableChat {
		game = append(game, "--disableChat")
	}
	if options.DisableMultiplayer {
		game = append(game, "--disableMultiplayer")
	}
	if runtime.GOOS == "darwin" {
		java = append(java, "-XstartOnFirstThread")
	}
	if options.Config.MinMemory != 0 {
		java = append(java, fmt.Sprintf("-Xms%dm", options.Config.MinMemory))
	}
	if options.Config.MaxMemory != 0 {
		java = append(java, fmt.Sprintf("-Xmx%dm", options.Config.MaxMemory))
	}

	for _, arg := range version.Arguments.Game {
		if arg, ok := arg.(string); ok {
			game = append(game, arg)
		}
	}
	for _, arg := range version.Arguments.Jvm {
		if arg, ok := arg.(string); ok {
			arg = strings.ReplaceAll(arg, "${version_name}", version.ID)
			arg = strings.ReplaceAll(arg, "${library_directory}", env.LibrariesDir)
			arg = strings.ReplaceAll(arg, "${classpath_separator}", string(os.PathListSeparator))
			java = append(java, arg)
		}
	}
	return java, game
}

// postProcess takes all Forge post processors and runs them with specified launch environment.
func postProcess(launchEnv LaunchEnvironment, processors []meta.ForgeProcessor) error {
	for _, processor := range processors {
		cmd := exec.Command(launchEnv.JavaPath, processor.JavaArgs...)
		cmd.Dir = launchEnv.GameDir
		cmd.Stderr = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// fetchVersion returns a VersionMeta containing both information for the base game, and specified mod loader.
func fetchVersion(loader Loader, gameVersion string, loaderVersion string) (meta.VersionMeta, error) {
	var loaderMeta meta.VersionMeta
	var err error

	version, err := meta.FetchVersionMeta(gameVersion)
	if err != nil {
		return meta.VersionMeta{}, fmt.Errorf("retrieve version metadata: %w", err)
	}

	if loader == LoaderFabric {
		loaderMeta, err = meta.Fabric.FetchMeta(version.ID, loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve fabric metadata: %w", err)
		}
	} else if loader == LoaderQuilt {
		loaderMeta, err = meta.Quilt.FetchMeta(version.ID, loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve quilt metadata: %w", err)
		}
	} else if loader == LoaderNeoForge {
		if loaderVersion == "latest" {
			loaderVersion, err = meta.FetchNeoforgeVersion(version.ID)
			if err != nil {
				return meta.VersionMeta{}, fmt.Errorf("retrieve neoforge version: %w", err)
			}
		}
		loaderMeta, _, err = meta.Neoforge.FetchMeta(loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve neoforge metadata: %w", err)
		}
	} else if loader == LoaderForge {
		if loaderVersion == "latest" {
			loaderVersion, err = meta.FetchForgeVersion(version.ID)
			if err != nil {
				return meta.VersionMeta{}, fmt.Errorf("retrieve forge version: %w", err)
			}
		}
		loaderMeta, _, err = meta.Forge.FetchMeta(loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve forge metadata: %w", err)
		}
	}

	return meta.MergeVersionMeta(version, loaderMeta), nil
}
