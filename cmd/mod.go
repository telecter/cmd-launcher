package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	util "github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/api"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v2"
)

func download(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		cli.ShowSubcommandHelpAndExit(ctx, 1)
	}
	name := ctx.Args().First()
	gameVersion := ctx.Args().Get(1)
	loader := ctx.Args().Get(2)

	if loader == "" {
		loader = "minecraft"
	}

	project, err := api.GetModrinthProject(ctx.Args().First())
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get project info: %s", err), 1)
	}

	version, err := api.GetModrinthProjVersion(name, gameVersion, loader)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get version info: %s", err), 1)
	}
	path := launcher.GetVersionDir(ctx.String("dir"), gameVersion)
	if project.ProjectType == "mod" {
		path = filepath.Join(path, "mods")
	} else if project.ProjectType == "resourcepack" {
		path = filepath.Join(path, "resourcepacks")
	}
	util.DownloadFile(version.Files[0].URL, filepath.Join(path, version.Files[0].Filename))
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
	modsDir := filepath.Join(launcher.GetVersionDir(ctx.String("dir"), ctx.Args().First()), "mods")
	if _, err := os.Stat(modsDir); errors.Is(err, fs.ErrNotExist) {
		return cli.Exit("no mods directory found", 1)
	}
	err := browser.OpenFile(modsDir)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to open mods directory: %s", err), 1)
	}
	return nil
}

var Mod = &cli.Command{
	Name:  "mod",
	Usage: "Add and manage mods and resource packs",
	Subcommands: []*cli.Command{
		{
			Name:      "download",
			Usage:     "Download mods (modrinth)",
			Args:      true,
			ArgsUsage: "<id> <instance> [fabric|quilt]",
			Action:    download,
		},
		{
			Name:      "info",
			Usage:     "Show info about mods (modrinth)",
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
