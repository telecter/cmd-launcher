package cmd

import (
	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/launcher"
)

type Start struct {
	ID          string `arg:"" name:"id" help:"Instance to launch"`
	Username    string `name:"username" help:"Set your username to the provided value (launches game in offline mode)" short:"u"`
	Server      string `name:"server" help:"Join a server immediately upon starting the game" short:"s"`
	Demo        bool   `name:"demo" help:"Start the game in demo mode"`
	DisableMP   bool   `name:"disable-mp" help:"Disable multiplayer"`
	DisableChat bool   `name:"disable-chat" help:"Disable chat"`
}

func (c *Start) Run(ctx *kong.Context) error {
	inst, err := launcher.GetInstance(c.ID)
	if err != nil {
		return err
	}
	if err := launcher.Launch(inst, launcher.LaunchOptions{
		QuickPlayServer: c.Server,
		OfflineMode:     c.Username != "",
		LoginData: auth.MinecraftLoginData{
			Username: c.Username,
		},
		Demo:               c.Demo,
		DisableMultiplayer: c.DisableMP,
		DisableChat:        c.DisableChat,
	},
	); err != nil {
		return err
	}
	return nil
}
