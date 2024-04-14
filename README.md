# cmd-launcher
A minimal command line Minecraft launcher.

- [Installation](#installation)
- [Usage](#usage)
  - [Selecting Versions](#selecting-versions)
  - [Online Mode](#online-mode)
  - [Mod Loaders](#mod-loaders)
## Installation
You need the [Deno](https://deno.com) JavaScript runtime to use this launcher.

To download and install the launcher, use this command.
```sh
deno install -A -n cmd-launcher https://raw.githubusercontent.com/telectr/cmd-launcher/VERSION/cli.ts
```
Replace `VERSION` with the version of the launcher you want to install.

## Usage
To start the latest version of Minecraft, just run the binary without any flags.

### Selecting Versions
To choose the version of the game to run, use the `launch <version>` subcommand.  
```sh
cmd-launcher launch 1.20.4
```

> [!IMPORTANT]
> This launcher is not tested for versions < 1.19. It may not work, mainly because this launcher does not yet support installing native libraries.

### Online Mode
This launcher supports online mode. Without any options, it will automatically attempt to use online mode. As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u, --username <username>` flag to set your offline mode username. This will automatically launch the game in offline mode.

### Mod Loaders
This launcher currently supports Fabric and Quilt. Use the `fabric` and `quilt` specifiers to use them.

```sh
cmd-launcher launch 1.20.4:fabric
```
