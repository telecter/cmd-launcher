<img src="icon.png" width="180">


# cmd-launcher

A minimal command line Minecraft launcher.

- [Installation](#installation)
  - [Building from source](#building-from-source)
- [Usage](#usage)
  - [Selecting Versions](#selecting-versions)
  - [Online Mode](#online-mode)
  - [Mod Loaders](#mod-loaders)

## Installation
Make sure you have [Go](https://go.dev) installed.

**To install, run:**
```sh
go install github.com/telecter/cmd-launcher@main
```
### Building from source

1. Clone this repository with `git`
2. In the source directory, run `go build .`


## Usage
Run the `help` subcommand to get the usage information.

### Creating an instance

To create a new instance, use the `create` subcommand.
You can use the `--loader, -l` flag to set the mod loader (only fabric at the moment.)  
Use the `--version, -v` flag to set the game version. If no value is supplied, the latest release is used. Acceptable values also include `release` or `snapshot` for the latest of either.
```sh
cmd-launcher create -v 1.21.5 -l fabric CoolInstance
```

> [!IMPORTANT]
> This launcher has not been tested for versions < 1.19. It may not work.

### Starting the Game
To start Minecraft, simply run the `start` subcommand followed by the instance ID.

```sh
cmd-launcher start CoolInstance
```

### Online Mode
Without any options, the launcher will automatically attempt to use online mode.
As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u, --username <username>` flag before the version
to set your username and the game will automatically launch in offline mode.
