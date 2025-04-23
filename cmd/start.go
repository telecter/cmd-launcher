package cmd

import (
	"context"
	"log"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

func start(ctx context.Context, c *cli.Command) error {
	var loginData auth.MinecraftLoginData
	// online mode
	if c.String("username") == "" {
		if !auth.IsLoggedIn() {
			log.Println("No account found, authenticating...")
		}
		var err error
		loginData, err = auth.LoginWithMicrosoft()
		if err != nil {
			return cli.Exit(err, 1)
		}

	} else {
		loginData = auth.MinecraftLoginData{
			Username: c.String("username"),
		}
	}

	gameVersion := c.Args().First()
	if gameVersion == "" {
		gameVersion = "latest"
	}
	if err := launcher.Launch(gameVersion, launcher.LaunchOptions{
		ModLoader: c.String("loader"),
	}, loginData); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

var Start = &cli.Command{
	Name:      "start",
	Usage:     "Start the game",
	ArgsUsage: "<version>",
	Action:    start,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Usage:   "Set your username to the provided value (launch game in offline mode)",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "loader",
			Usage:   "Set the mod loader to use",
			Aliases: []string{"l"},
		},
	},
}
