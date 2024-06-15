package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/api"
	"github.com/urfave/cli/v2"
)

func add(ctx *cli.Context) error {
	if ctx.Args().Len() < 3 {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}
	name := ctx.Args().First()
	gameVersion := ctx.Args().Get(1)
	loader := ctx.Args().Get(2)
	err := api.DownloadModrinthProject(fmt.Sprintf("%s/instances/%s/mods", ctx.String("dir"), gameVersion), name, gameVersion, loader)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to download mod: %s", err), 1)
	}
	return nil
}
func info(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}
	project, err := api.GetModrinthProject(ctx.Args().First())
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get mod info: %s", err), 1)
	}
	fmt.Printf("%s\n%s\n", project.Title, project.Description)
	fmt.Printf("Latest: %s\n", project.GameVersions[len(project.GameVersions)-1])
	return nil
}
func show(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}
	modsDirectory := fmt.Sprintf("%s/instances/%s/mods", ctx.String("dir"), ctx.Args().First())
	if _, err := os.Stat(modsDirectory); errors.Is(err, fs.ErrNotExist) {
		return cli.Exit("no mods directory found", 1)
	}
	err := browser.OpenFile(modsDirectory)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to open mods directory: %s", err), 1)
	}
	return nil
}

var Mod = &cli.Command{
	Name:  "mod",
	Usage: "Add and manage mods (modrinth)",
	Subcommands: []*cli.Command{
		{
			Name:      "add",
			Usage:     "Add mods",
			Args:      true,
			ArgsUsage: "<id> <instance> <fabric|quilt>",
			Action:    add,
		},
		{
			Name:      "info",
			Usage:     "Show info about mods",
			Args:      true,
			ArgsUsage: "<id>",
			Action:    info,
		},
		{
			Name:      "show",
			Usage:     "Open mods directory for the specified instance",
			Args:      true,
			ArgsUsage: "<instance>",
			Action:    show,
		},
	},
}
