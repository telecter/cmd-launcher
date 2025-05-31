package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/launcher"
)

type Delete struct {
	ID string `arg:"" name:"id" help:"Instance to delete"`
}

func (c *Delete) Run(ctx *kong.Context) error {
	inst, err := launcher.GetInstance(c.ID)
	if err != nil {
		return err
	}
	var input string
	fmt.Printf("Are you sure you want to delete '%s'?\nIt will be gone forever (a long time!) [y/n] ", inst.Name)
	fmt.Scanln(&input)

	if input == "y" || input == "Y" {
		if err := launcher.DeleteInstance(c.ID); err != nil {
			return fmt.Errorf("remove instance: %w", err)
		}
		fmt.Printf("Deleted instance '%s'\n", inst.Name)
	} else {
		fmt.Println("Operation aborted")
	}
	return nil
}
