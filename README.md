<img src="docs/icon.png" width="180">


# cmd-launcher

A minimal command line Minecraft launcher.

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

**To install, run:**
```bash
go install github.com/telecter/cmd-launcher@latest
```
### Building from source

1. Clone this repository with `git`
2. In the source directory, run `go build .`


## Usage
Use the `--help` flag to get the usage information.

### Creating an instance

To create a new instance, use the `create` command.  
You can use the `--loader, -l` flag to set the mod loader. Only Fabric and Quilt are supported at the moment. 

Use the `--version, -v` flag to set the game version. If no value is supplied, the latest release is used. Acceptable values also include `release` or `snapshot` for the latest of either.
```sh
cmd-launcher create -v 1.21.5 -l fabric CoolInstance
```
If you want to delete an instance, use the `delete` command.
> [!IMPORTANT]
> This launcher has not been tested for versions < 1.14. It may not work, but I am working on fixing these issues.

### Starting the Game
To start Minecraft, simply run the `start` command followed by the instance ID.

```bash
cmd-launcher start CoolInstance
```

### Authentication
If you want to play the game in online mode, you will need to add a Microsoft account. To do this, use the `auth login` command. As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication. This can be avoided with the `--no-browser` flag. The launcher will automatically attempt to start the game in online mode if there is an account present.


To play in offline mode, just pass the `-u, --username <username>` flag to the `start` command
to set your username and the game will automatically launch in offline mode.


Log out via the `auth logout` command.
### Instance Configuration
To change configuration values for an instance, go to the instance directory and open the `instance.json` file.

Currently configurable values are:
* Window Resolution
* Java Executable Path
* Minimum and maximum memory

These values can be overriden when starting the game via command line flags in the `start` command.

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
    "min_memory": 512,
    "max_memory": 4096
  }
}
```
### Search
The `search` command can search for Minecraft versions or instances that you have. While it defaults to searching for installed instances, it can also search for game, Fabric, and Quilt versions.

```bash
cmd-launcher search [<query>] [--kind {instances, versions, fabric, quilt}]
```