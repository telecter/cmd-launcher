# cmd-launcher
A minimal command line Minecraft launcher.

```javascript
./cmd-launcher
[...]
Starting Minecraft 1.20.4...

```
## Installation
You can find releases of this launcher in the Releases section.

### Running From Source
Since this launcher uses TypeScript, its pretty simple to run from source.  

1. Make sure you have the [Deno](https://deno.com) JS runtime installed
2. Download the source code from this repository
3. Run `main.ts` with `deno run -A main.ts`

> [!NOTE]
> Since the binaries effectively contain the Deno runtime, they are usually around 60 MB. The code itself is only a few kilobytes.

## Usage
To start the latest version of Minecraft, just run the binary without any flags.

### Selecting Versions
To choose the version of the game to run, pass the `-l, --launch <version>` flag.  
```crystal
cmdlauncher --launch 1.20.4
```

> [!IMPORTANT]
> This launcher is not tested for versions < 1.19. It may not work, mainly because this launcher does not yet support installing native libraries.

### Online Mode
This launcher supports online mode. Without any options, it will automatically attempt to use online mode. As part of Microsoft's OAuth2 flow, the default web browser will be opened to complete the authentication.

If you want to use offline mode, just pass the `-u <username>` flag to set your offline mode username. This will automatically launch the game in offline mode.
