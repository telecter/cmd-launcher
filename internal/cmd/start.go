package cmd

import (
	"fmt"
	"os"
	"time"

	"dario.cat/mergo"
	"github.com/alecthomas/kong"
	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/pkg/auth"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

type watcher struct {
	DownloadProgressBar *progressbar.ProgressBar
	Verbose             bool
}

func (watcher watcher) Handle(event any) {
	switch e := event.(type) {
	case launcher.DownloadingEvent:
		watcher.DownloadProgressBar.ChangeMax(e.Total)
		watcher.DownloadProgressBar.Add(1)
	case launcher.AssetsResolvedEvent:
		if watcher.Verbose {
			fmt.Printf("Identified %d assets\n", e.Assets)
		}
	case launcher.LibrariesResolvedEvent:
		if watcher.Verbose {
			fmt.Printf("Identified %d libraries\n", e.Libraries)
		}
	case launcher.MetadataResolvedEvent:
		if watcher.Verbose {
			fmt.Println("Version metadata retrieved")
		}
	}
}

type Start struct {
	ID      string `arg:"" name:"id" help:"Instance to launch"`
	Verbose bool   `name:"verbose" help:"Increase verbosity" short:"v"`

	Username    string `name:"username" help:"Set your username to the provided value (launches game in offline mode)" short:"u" group:"Game Options"`
	Server      string `name:"server" help:"Join a server immediately upon starting the game" short:"s" group:"Game Options"`
	Demo        bool   `name:"demo" help:"Start the game in demo mode" group:"Game Options"`
	DisableMP   bool   `name:"disable-mp" help:"Disable multiplayer" group:"Game Options"`
	DisableChat bool   `name:"disable-chat" help:"Disable chat" group:"Game Options"`
	Width       int    `name:"width" help:"Game window width" group:"Configuration Overrides"`
	Height      int    `name:"height" help:"Game window height" group:"Configuration Overrides" `
	JVM         string `name:"jvm" help:"Path to the JVM to use" group:"Configuration Overrides" type:"path" placeholder:"PATH"`
	MinMemory   int    `name:"min-memory" help:"Minimum memory" group:"Configuration Overrides" placeholder:"MB"`
	MaxMemory   int    `name:"max-memory" help:"Maximum memory" group:"Configuration Overrides" placeholder:"MB"`
}

func (c *Start) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}

	config := inst.Config
	override := launcher.InstanceConfig{
		WindowResolution: struct {
			Width  int "json:\"width\""
			Height int "json:\"height\""
		}{
			Width:  c.Width,
			Height: c.Height,
		},
		Java:      c.JVM,
		MinMemory: c.MinMemory,
		MaxMemory: c.MaxMemory,
	}
	mergo.Merge(&config, override, mergo.WithOverride)

	session := auth.Session{
		Username: c.Username,
	}
	if c.Username == "" {
		session, err = auth.Authenticate()
		if err != nil {
			return fmt.Errorf("authenticate session: %w", err)
		}
	}
	options := launcher.EnvOptions{
		Session:            session,
		Config:             config,
		QuickPlayServer:    c.Server,
		Demo:               c.Demo,
		DisableMultiplayer: c.DisableMP,
		DisableChat:        c.DisableChat,
	}

	launchEnv, err := inst.Prepare(options, watcher{
		DownloadProgressBar: progressbar.NewOptions(0,
			progressbar.OptionSetDescription("Downloading files"),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionOnCompletion(func() {
				fmt.Print("\n")
			}),
			progressbar.OptionFullWidth(),
		),
		Verbose: c.Verbose,
	})

	if err != nil {
		return err
	}
	return launchEnv.Launch(launcher.ConsoleRunner{})
}
