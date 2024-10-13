package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"codeberg.org/telecter/cmd-launcher/internal/auth"
	"codeberg.org/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v2"
)

func start(ctx *cli.Context) error {
	var loginData auth.MinecraftLoginData
	// online mode
	if ctx.String("username") == "" {
		accountDataPath := filepath.Join(ctx.String("dir"), "account.txt")
		var refresh string
		data, err := os.ReadFile(accountDataPath)
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("no account found, authenticating...")
			refresh = ""
		} else {
			refresh = string(data)
		}
		loginData, err = auth.LoginWithMicrosoft(refresh)
		if err != nil {
			return cli.Exit(err, 1)
		}
		os.WriteFile(accountDataPath, []byte(loginData.Refresh), 0644)
	} else {
		loginData = auth.MinecraftLoginData{
			Username: ctx.String("username"),
		}
	}
	if err := launcher.Launch(ctx.Args().Get(0), ctx.String("dir"), launcher.LaunchOptions{
		ModLoader: ctx.String("loader"),
	}, loginData); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

var Start = &cli.Command{
	Name:      "start",
	Usage:     "Start the game",
	Args:      true,
	ArgsUsage: " <version>",
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
