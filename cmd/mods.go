package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/urfave/cli/v3"
)

var searchCommand = &cli.Command{
	Name:  "search",
	Usage: "Search mods on Modrinth",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name: "query",
		},
	},
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "page",
			Usage:   "Results page to show",
			Aliases: []string{"p"},
			Value:   1,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		results, err := meta.SearchModrinthProjects(c.StringArg("query"), c.Int("page"))
		if err != nil {
			return cli.Exit(fmt.Errorf("failed to search for projects: %w", err), 1)
		}
		fmt.Printf("Page %d", c.Int("page"))
		for _, hit := range results.Hits {
			fmt.Printf("\n%s by %s\n", hit.Title, hit.Author)
			fmt.Println(strings.Join(hit.DisplayCategories, ", "))
			fmt.Println(hit.Description)
			fmt.Printf("%d downloads • %d followers\n", hit.Downloads, hit.Follows)
		}
		return nil
	},
}

var infoCommand = &cli.Command{
	Name:  "info",
	Usage: "Show Modrinth mod information",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name: "id",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if c.StringArg("id") == "" {
			cli.ShowSubcommandHelpAndExit(c, 1)
		}
		project, err := meta.GetModrinthProject(c.StringArg("id"))
		if err != nil {
			return cli.Exit(fmt.Errorf("failed to get project: %w", err), 1)
		}
		versions, err := meta.GetModrinthProjectVersions(project.ID)
		if err != nil {
			return cli.Exit(fmt.Errorf("failed to get project versions: %w", err), 1)
		}
		fmt.Printf("%s\n", project.Title)
		fmt.Printf("%d downloads • %d followers\n", project.Downloads, project.Followers)
		fmt.Println("-------------------")
		fmt.Println(strings.Join(project.Categories, ", "))
		fmt.Println(project.Description)

		fmt.Println("\nLatest Version:")
		fmt.Println(versions[0].Name)
		fmt.Println(strings.Join(versions[0].GameVersions, ", "))
		fmt.Println(strings.Join(versions[0].Loaders, ", "))
		return nil
	},
}

var ModsCommand = &cli.Command{
	Name:  "mods",
	Usage: "Manage mods",
	Commands: []*cli.Command{
		searchCommand,
		infoCommand,
	},
}
