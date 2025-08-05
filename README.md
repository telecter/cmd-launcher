<img src="docs/icon.png" width="180">


# cmd-launcher

A minimal command line Minecraft launcher.

[![Build](https://github.com/telecter/cmd-launcher/actions/workflows/build.yml/badge.svg)](https://github.com/telecter/cmd-launcher/actions/workflows/build.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/telecter/cmd-launcher)
[![Go Reference](https://pkg.go.dev/badge/github.com/telecter/cmd-launcher.svg)](https://pkg.go.dev/github.com/telecter/cmd-launcher)

- [Installation](#installation)
  - [Building from source](#building-from-source)
- [Usage](#usage)
  - [Creating an instance](#creating-an-instance)
  - [Starting the game](#starting-the-game)
  - [Authentication](#authentication)
  - [Instance Configuration](#instance-configuration)
  - [Search](#search)

[API Documentation](docs/API.md)

## Installation
Make sure you have [Go](https://go.dev) installed.

**To install the latest version, run:**
```bash
go install github.com/telecter/cmd-launcher@latest
```
Replace `latest` with `main` for the latest commit.
### Building from source

1. Clone the repository: `git clone https://github.com/telecter/cmd-launcher`
2. In the source directory, run `go run .` to compile and run the launcher.
3. Once you are ready, compile the executable with `go build .`


## Usage
Use the `--help` flag to get the usage information of any command.

### Instances
**Creating an instance**  
To create a new instance, use the `inst create` command.  
You can use the `--loader, -l` flag to set the mod loader. Forge, NeoForge, Fabric, and Quilt are all supported. If you want to select a specific version of the loader, use the `--loader-version` flag. Otherwise, the latest applicable version is chosen.

Use the `--version, -v` flag to set the game version. If no value is supplied, the latest release is used. Acceptable values also include `release` or `snapshot` for the latest of either.

When starting the game, the launcher will attempt to download a Java runtime from Mojang. If it can't find a suitable one, you will need to set one manually in the instance configuration.
```sh
cmd-launcher inst create -v 1.21.5 -l fabric CoolInstance
```

**Deleting instances**  
If you want to delete an instance, use the `inst delete` command followed by the instance name.

### Starting the Game
> [!IMPORTANT]
> This launcher has not been tested for versions < 1.14. It may not work, but I am working on fixing these issues.

To start Minecraft, simply run the `start` command followed by the name of the instance you want to start.

```bash
cmd-launcher start CoolInstance
```
To set game options and override instance configuration, you can set specific flags on the `start` command. These can be viewed in the help text.

**Verbosity**  
To increase the verbosity of the launcher, use the `--verbosity` flag. It can be set to either:
* `info` - default, no extra logging
* `extra` - more information when starting the game
* `debug` - debug information useful for debugging the launcher 

### Authentication
If you want to play the game in online mode, you will need to add a Microsoft account.

 To do this, use the `auth login` command. As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication. This can be avoided with the `--no-browser` flag.  
The launcher will automatically attempt to start the game in online mode if there is an account present.


To play in offline mode, just pass the `-u, --username <username>` flag to the `start` command
to set your username and the game will automatically launch in offline mode.


You can log out via the `auth logout` command.
### Instance Configuration
To change configuration values for an instance, navigate to the instance directory and open the `instance.json` file.

Configurable values are:
* Game version
* Mod loader and version (if not vanilla)
* Window resolution
* Java executable path (if empty, Mojang Java will be downloaded)
* Custom JAR path to use instead of downloading the normal client JAR
* Extra Java args
* Minimum and maximum memory

As mentioned previously, these values can be overriden with command line flags.

**Example `instance.json` file**
```json
{
  "game_version": "1.21.5",
  "mod_loader": "fabric",
  "mod_loader_version": "0.16.14",
  "config": {
    "resolution": {
      "width": 1708,
      "height": 960
    },
    "java": "/usr/bin/java",
    "java_args": "",
    "custom_jar": "",
    "min_memory": 512,
    "max_memory": 4096
  }
}
```
### Search
The `search` command can search for Minecraft  or mod loader versions. It defaults to searching for game versions, but can also be used to search for Fabric and Quilt versions.

```bash
cmd-launcher search [<query>] [--kind {versions, fabric, quilt}]
```