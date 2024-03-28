# cmd-launcher

A minimal command line Minecraft launcher.

```javascript
./cmd-launcher
[...]
Starting Minecraft 1.20.4...

```

## Installation
You can find releases of this launcher in the Releases section.


Since this launcher uses JavaScript with the Deno runtime, its pretty simple to run from source.  

1. Make sure you have the [Deno](https://deno.com) runtime installed
2. Download the source code from this repository
3. Run `main.ts` with `deno run -A main.ts`

> **Note:** Since the binaries effectively contain the Deno runtime, they are usually around 60 MB. The code itself is only a few kilobytes.

## Usage

### Starting Minecraft
To start the latest version of Minecraft, just run the binary without any flags.  
If you want to change the version, pass the `-l <version>` flag to select it.

```
usage: cmd-launcher [...options]
A command line Minecraft launcher.

Options:
-l, --launch      Launch a specific version of the game
-u, --username    Set the offline mode username
--update          Update the launcher
-h, --help        Show this help and exit
```

### Online Mode
This launcher supports online mode. Without any options, it will automatically attempt to use online mode. As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u <username>` flag to set your offline mode username. This will automatically launch the game in offline mode.
