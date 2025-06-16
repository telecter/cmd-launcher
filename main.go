package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/internal/cli/cmd"

	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const (
	name        = "cmd-launcher"
	description = "A minimal command line Minecraft launcher."
	version     = "1.1.0"
	license     = "Licensed MIT"
	copyright   = "Copyright (c) 2024-2025 telecter"
)

type about struct{}

func (about) Run(ctx *kong.Context) error {
	color.New(color.Bold).Printf("%s %s\n", name, version)
	color.New(color.Underline).Println(description)
	fmt.Println(copyright)
	fmt.Println(license)
	return nil
}

type CLI struct {
	Start     cmd.Start    `cmd:"" help:"${cmd_start}"`
	Instance  cmd.Instance `cmd:"" help:"${cmd_instance}" aliases:"inst"`
	Auth      cmd.Auth     `cmd:"" help:"${cmd_auth}"`
	Search    cmd.Search   `cmd:"" help:"${cmd_search}"`
	About     about        `cmd:"" help:"${cmd_about}"`
	Verbosity string       `help:"${verbosity}" enum:"info,extra,debug" default:"info"`
	Dir       string       `help:"${dir}" type:"path" placeholder:"PATH"`
	NoColor   bool         `help:"${nocolor}"`
}

func (c *CLI) AfterApply(ctx *kong.Context) error {
	var verbosity int
	switch c.Verbosity {
	case "info":
		verbosity = 0
	case "extra":
		verbosity = 1
	case "debug":
		verbosity = 2
	}
	ctx.Bind(verbosity)
	if c.Dir != "" {
		if err := env.SetDirs(c.Dir); err != nil {
			return err
		}
	}
	if err := auth.ReadFromCache(); err != nil {
		return fmt.Errorf("read auth store: %w", err)
	}
	if c.NoColor {
		color.NoColor = true
	}
	return nil
}

func main() {
	vars := make(kong.Vars)
	for k, v := range cli.Translations() {
		vars[strings.ReplaceAll(k, ".", "_")] = v
	}

	parser := kong.Must(&CLI{},
		kong.UsageOnError(),
		kong.Name(name),
		kong.Description(description),
		kong.ConfigureHelp(kong.HelpOptions{
			NoExpandSubcommands: true,
		}),
		kong.ValueFormatter(func(value *kong.Value) string {
			if value.Enum != "" {
				return fmt.Sprintf("%s [%s]", value.Help, strings.Join(value.EnumSlice(), ", "))
			}
			return value.Help
		}),
		kong.Groups{
			"overrides": cli.Translate("cmd.start.overrides"),
			"opts":      cli.Translate("cmd.start.opts"),
		},
		vars,
	)

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		exitCode := 1
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			parseErr.Context.PrintUsage(false)
			exitCode = parseErr.ExitCode()
		}
		cli.Error("%s", err)
		parser.Exit(exitCode)
	}

	if err := ctx.Run(); err != nil {
		cli.Error("%s", err)
		var coder kong.ExitCoder
		if errors.As(err, &coder) {
			ctx.Exit(coder.ExitCode())
		}
		ctx.Exit(1)
	}
}
