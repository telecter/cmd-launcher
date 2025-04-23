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

### Selecting Versions

To choose the version of the game to run, use the `start` subcommand.

```sh
cmd-launcher start 1.21.5
```
If you do not provide a version, the lastest version will be automatically chosen.

> [!IMPORTANT]
> This launcher has not been tested for versions < 1.19. It may not work.

### Online Mode
Without any options, the launcher will automatically attempt to use online mode.
As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u, --username <username>` flag before the version
to set your username and the game will automatically launch in offline mode.

### Mod Loaders

Currently, only Fabric is supported. Use the `-l, --loader` flag to use them.

```sh
cmd-launcher start -l fabric 1.21.5
```
