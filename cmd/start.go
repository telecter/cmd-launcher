package cmd

import (
	"context"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

func start(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}

	if err := launcher.Launch(c.Args().First(), launcher.LaunchOptions{
		QuickPlayServer: c.String("server"),
		OfflineMode:     c.String("username") != "",
		LoginData: auth.MinecraftLoginData{
			Username: c.String("username"),
		},
	},
	); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

var Start = &cli.Command{
	Name:      "start",
	Usage:     "Start the specified instance",
	ArgsUsage: "<id>",
	Action:    start,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Usage:   "Sets your username to the provided value (launches game in offline mode)",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "server",
			Usage:   "Joins a server immediately upon starting the game.",
			Aliases: []string{"s"},
		},
	},
}
