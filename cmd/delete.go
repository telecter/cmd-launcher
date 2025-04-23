package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

func delete(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}

	instance, err := launcher.GetInstance(c.Args().First())
	if err != nil {
		return cli.Exit(err, 1)
	}
	var input string
	fmt.Printf("Are you sure you want to delete '%s'?\nIt will be gone forever (a long time!) [y/n] ", instance.Name)
	fmt.Scanln(&input)

	if input == "y" {
		if err := launcher.DeleteInstance(c.Args().First()); err != nil {
			return cli.Exit(fmt.Errorf("failed to remove instance: %w", err), 1)
		}
	} else {
		log.Println("Operation aborted.")
	}
	return nil
}

var Delete = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a Minecraft instance",
	ArgsUsage: "<id>",
	Action:    delete,
}
