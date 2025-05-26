package cmd

import (
	"context"
	"fmt"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

var Create = &cli.Command{
	Name:  "create",
	Usage: "Create a new Minecraft instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "loader",
			Usage:   "Mod loader to use",
			Aliases: []string{"l"},
			Value:   "vanilla",
		},
		&cli.StringFlag{
			Name:    "version",
			Usage:   "Game version",
			Aliases: []string{"v"},
			Value:   "release",
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

		var loader launcher.Loader
		switch c.String("loader") {
		case launcher.LoaderFabric.String():
			loader = launcher.LoaderFabric
		case launcher.LoaderQuilt.String():
			loader = launcher.LoaderQuilt
		case launcher.LoaderVanilla.String():
			loader = launcher.LoaderVanilla
		default:
			return fmt.Errorf("invalid mod loader")
		}

		inst, err := launcher.CreateInstance(launcher.InstanceOptions{
			GameVersion: c.String("version"),
			Name:        c.StringArg("id"),
			Loader:      loader,
		})
		if err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}
		fmt.Printf("Created instance '%s' with Minecraft %s ", inst.Name, inst.GameVersion)
		if inst.Loader != launcher.LoaderVanilla {
			fmt.Printf("(%s %s)\n", inst.Loader, inst.LoaderVersion)
		}
		return nil
	},
}
