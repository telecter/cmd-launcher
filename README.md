# cmd-launcher

A minimal command line Minecraft launcher.

- [Installation](#installation)
- [Usage](#usage)
  - [Selecting Versions](#selecting-versions)
  - [Online Mode](#online-mode)
  - [Mod Loaders](#mod-loaders)

## Installation

Make sure you have [Go](https://go.dev) installed.

1. Clone this repository with `git`
2. `cd` into the directory and run `go build`
3. Run the binary that is created

> Prebuilt binaries will be available in the future.

## Usage
Run the `help` subcommand to get the usage information.

### Selecting Versions

To choose the version of the game to run, use the `launch <version>` subcommand.

```sh
cmd-launcher launch 1.20.6
```

> [!IMPORTANT]
> This launcher is not tested for versions < 1.19. It may not work,
> mainly because this launcher does not yet support installing native libraries.

### Online Mode
**Online mode will not work right now** due to issues with Microsoft Azure.


Without any options, the launcher will automatically attempt to use online mode.
As part of Microsoft's OAuth2 flow, the default web browser will be opened to
complete the authentication.

If you want to use offline mode, just pass the `-u, --username <username>` flag before the version
to set your username and the game will automatically launch in offline mode.

### Mod Loaders

Currently, only Fabric and Quilt are supported. Use the `-l, --loader` flags to use them.

```sh
cmd-launcher -l fabric 1.20.6
```
