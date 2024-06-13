package cmd

import (
	"fmt"

	"github.com/telecter/cmd-launcher/pkg/api"
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

var Mod = &cli.Command{
	Name:  "mod",
	Usage: "Add and manage mods (modrinth)",
	Subcommands: []*cli.Command{
		{
			Name:      "add",
			Usage:     "Add mods",
			Args:      true,
			ArgsUsage: "<id> <gameVersion> <fabric|quilt>",
			Action:    add,
		},
		{
			Name:      "info",
			Usage:     "Show info about mods",
			Args:      true,
			ArgsUsage: "<id>",
			Action:    info,
		},
	},
}
