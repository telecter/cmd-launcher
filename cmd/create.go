package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

func create(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}

	if _, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion: "release",
		Name:        c.Args().First(),
		ModLoader:   c.String("loader"),
	}); err != nil {
		return cli.Exit(fmt.Errorf("failed to create instance: %w", err), 1)
	}
	log.Println("Created new instance.")
	return nil
}

var Create = &cli.Command{
	Name:      "create",
	Usage:     "Create a new Minecraft instance",
	ArgsUsage: "<id>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "loader",
			Usage:   "Mod loader to use",
			Aliases: []string{"l"},
		},
	},
	Action: create,
}
