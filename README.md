<img src="icon.png" width="180">


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
  - [Caching](#caching)

## Installation
Make sure you have [Go](https://go.dev) installed.

**To install, run:**
```bash
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
If you want to delete an instance, use the `delete` subcommand.
> [!IMPORTANT]
> This launcher has not been tested for versions < 1.19. It may not work, but I am working on fixing these issues.

### Starting the Game
To start Minecraft, simply run the `start` subcommand followed by the instance ID.

```bash
cmd-launcher start CoolInstance
```

### Authentication
Without any options, the launcher will automatically attempt to use online mode.
As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u, --username <username>` flag before the version
to set your username and the game will automatically launch in offline mode.

### Instance Configuration
To change configuration values for an instance, go to the instance directory and open the `instance.json` file.

Currently configurable values are:
* Window Resolution
* Java Executable Path
* Minimum and maximum memory

**Example `instance.json` file**
```json
{
  "dir": "<path of instance directory>",
  "game_version": "1.21.5",
  "name": "CoolInstance",
  "mod_loader": "fabric",
  "config": {
    "window_resolution": [
      1708,
      960
    ],
    "java_location": "/usr/bin/java",
    "min_memory": 512,
    "max_memory": 4096
  }
}
```
### Search
The `search` subcommand can search for Minecraft versions or instances that you have.

```bash
cmd-launcher search <versions|instances> [query]
```

### Caching
Because the launcher tries to use cached metadata when possible, you may need to include the `--clear-caches` flag once to get it to download new metadata.