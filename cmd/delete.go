package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

var Delete = &cli.Command{
	Name:  "delete",
	Usage: "Delete a Minecraft instance",
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
		var input string
		fmt.Printf("Are you sure you want to delete '%s'?\nIt will be gone forever (a long time!) [y/n] ", inst.Name)
		fmt.Scanln(&input)

		if input == "y" {
			if err := launcher.DeleteInstance(c.StringArg("id")); err != nil {
				return fmt.Errorf("failed to remove instance: %w", err)
			}
			log.Printf("Deleted instance '%s'\n", inst.Name)
		} else {
			log.Println("Operation aborted.")
		}
		return nil
	},
}
