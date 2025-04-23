package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/telecter/cmd-launcher/internal/network"
	"github.com/telecter/cmd-launcher/internal/network/api"
	"github.com/urfave/cli/v3"
)

func download(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 2 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}
	name := c.Args().First()
	gameVersion := c.Args().Get(1)
	loader := c.Args().Get(2)

	if loader == "" {
		loader = "minecraft"
	}

	project, err := api.GetModrinthProject(c.Args().First())
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get project info: %s", err), 1)
	}

	version, err := api.GetModrinthProjVersion(name, gameVersion, loader)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get version info: %s", err), 1)
	}
	path := launcher.GetVersionDir(c.String("dir"), gameVersion)
	if project.ProjectType == "mod" {
		path = filepath.Join(path, "mods")
	} else if project.ProjectType == "resourcepack" {
		path = filepath.Join(path, "resourcepacks")
	}
	network.DownloadFile(version.Files[0].URL, filepath.Join(path, version.Files[0].Filename))
	return nil
}
func info(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}
	project, err := api.GetModrinthProject(c.Args().First())
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get mod info: %s", err), 1)
	}
	fmt.Printf("%s\n%s\n", project.Title, project.Description)
	fmt.Printf("Latest: %s\n", project.GameVersions[len(project.GameVersions)-1])
	return nil
}
func show(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}
	modsDir := filepath.Join(launcher.GetVersionDir(c.String("dir"), c.Args().First()), "mods")
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
	Usage: "Add and manage mods and resource packs from Modrinth",
	Commands: []*cli.Command{
		{
			Name:      "download",
			Usage:     "Download mods",
			ArgsUsage: "<id> <instance> [fabric|quilt]",
			Action:    download,
		},
		{
			Name:      "info",
			Usage:     "Show info about mods",
			ArgsUsage: "<id>",
			Action:    info,
		},
		{
			Name:      "show",
			Usage:     "Open mods directory for the specified instance",
			ArgsUsage: "<instance>",
			Action:    show,
		},
	},
}
