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
}

func (watcher watcher) Handle(event any) {
	switch e := event.(type) {
	case launcher.DownloadingEvent:
		watcher.DownloadProgressBar.ChangeMax(e.Total)
		watcher.DownloadProgressBar.Add(1)
	}
}

type Start struct {
	ID          string `arg:"" name:"id" help:"Instance to launch"`
	Username    string `name:"username" help:"Set your username to the provided value (launches game in offline mode)" short:"u"`
	Server      string `name:"server" help:"Join a server immediately upon starting the game" short:"s"`
	Demo        bool   `name:"demo" help:"Start the game in demo mode"`
	DisableMP   bool   `name:"disable-mp" help:"Disable multiplayer"`
	DisableChat bool   `name:"disable-chat" help:"Disable chat"`
	Width       int    `name:"width" help:"Game window width" group:"Configuration Overrides"`
	Height      int    `name:"height" help:"Game window height" group:"Configuration Overrides" `
	JVM         string `name:"jvm" help:"Path to the JVM to use" group:"Configuration Overrides" type:"path" placeholder:"PATH"`
	MinMemory   int    `name:"min-memory" help:"Minimum memory" group:"Configuration Overrides" placeholder:"MB"`
	MaxMemory   int    `name:"max-memory" help:"Maximum memory" group:"Configuration Overrides" placeholder:"MB"`
}

func (c *Start) Run(ctx *kong.Context) error {
	inst, err := launcher.GetInstance(c.ID)
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
		session, err = auth.Authenticate(clientID)
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

	launchEnv, err := launcher.Prepare(inst, options, watcher{
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
		)})

	if err != nil {
		return err
	}
	return launcher.Launch(launchEnv, launcher.ConsoleRunner{})
}
