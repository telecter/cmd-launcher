package cmd

import (
	"context"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

var Start = &cli.Command{
	Name:  "start",
	Usage: "Start the specified instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Usage:   "Set your username to the provided value (launches game in offline mode)",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "server",
			Usage:   "Join a server immediately upon starting the game.",
			Aliases: []string{"s"},
		},
		&cli.BoolFlag{
			Name:  "demo",
			Usage: "Start the game in demo mode",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "disable-mp",
			Usage: "Disable multiplayer",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "disable-chat",
			Usage: "Disable chat",
			Value: false,
		},
	},
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name: "id",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if c.StringArg("id") == "" {
			cli.ShowSubcommandHelpAndExit(c, 1)
		}

		inst, err := launcher.GetInstance(c.StringArg("id"))
		if err != nil {
			return err
		}
		if err := launcher.Launch(inst, launcher.LaunchOptions{
			QuickPlayServer: c.String("server"),
			OfflineMode:     c.String("username") != "",
			LoginData: auth.MinecraftLoginData{
				Username: c.String("username"),
			},
			Demo:               c.Bool("demo"),
			DisableMultiplayer: c.Bool("disable-mp"),
			DisableChat:        c.Bool("disable-chat"),
		},
		); err != nil {
			return err
		}
		return nil
	},
}
