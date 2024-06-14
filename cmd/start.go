package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/telecter/cmd-launcher/pkg/api"
	"github.com/telecter/cmd-launcher/pkg/launcher"
	"github.com/urfave/cli/v2"
)

func start(ctx *cli.Context) error {
	var authData api.AuthData
	// online mode
	if ctx.String("username") == "" {
		var refresh string
		data, err := os.ReadFile(ctx.String("dir") + "/account.txt")
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("no account data file found")
			refresh = ""
		} else {
			refresh = string(data)
		}
		authData, err = api.GetAuthData(refresh)
		if err != nil {
			return cli.Exit(err, 1)
		}
	} else { // offline mode (TODO: don't fill in auth data with fake info, maybe actually fix the problem?)
		authData = api.AuthData{
			Token:    "not a real token",
			UUID:     "bbccc3e6-9f7f-4270-a3f0-4d126765523e",
			Username: ctx.String("username"),
			Refresh:  "fake refresh",
		}
	}
	if err := launcher.Launch(ctx.Args().Get(0), ctx.String("dir"), launcher.LaunchOptions{
		ModLoader: ctx.String("loader"),
	}, authData); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

var Start = &cli.Command{
	Name:      "start",
	Usage:     "Start the game",
	Args:      true,
	ArgsUsage: " [version]",
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
