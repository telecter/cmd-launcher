import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.207.0/cli/parse_args.ts";
import * as api from "./api.ts"
import * as util from "./util.ts"


async function update() {
  if (Deno.execPath().includes("deno")) {
    throw Error("Cannot update non-executable")
  }
  const tags = await (await fetch("https://api.github.com/repos/telectr/minecraft-launcher/tags")).json()
  const latestTag = tags[0].name
  console.log(`Upgrading to ${latestTag}`)
  const data = await (await fetch(`https://github.com/telectr/minecraft-launcher/releases/download/${latestTag}/launcher-${Deno.build.target}`)).arrayBuffer()
  await Deno.writeFile(Deno.execPath(), new Uint8Array(data))
  Deno.exit()
}

function getArgs(args: string[]) {
  const flags = parseArgs(args, {
    string: ["version", "username", "list"],
    boolean: ["help", "update"],
    alias: {
      "version": "v",
      "username": "u",
      "list": "l"
    }
  })
  if (flags.help) {
    console.log(`
Command Line Minecraft Launcher
Usage: minecraft-launcher [options]
-v, --version         Version of Minecraft to launch
-u, --username        Username for offline mode, defaults to random
-l, --list <type>     List all versions of the specified type
--update              Update the launcher
--help                Display this help and exit
    `)
    Deno.exit()
  }
  if (flags._.length > 0) {
    console.log("Invalid argument")
    Deno.exit(1)
  }
  return flags
}

async function main(args: string[]) {
  const flags = getArgs(args)
  let version: string|null = null
  if (flags.version) {
    version = flags.version
  }

  if (flags.update) {
    await update()
  }

  if (flags.list) {
    try {
      (await api.filterVersions(flags.list)).forEach((element) => console.log(element.id))
    }
    catch (err) {
      console.log(`Failed to list versions: ${err.message}`)
      Deno.exit(1)
    }
    Deno.exit()
  }

  console.log(`Getting version data for ${version ?? "latest"}`)
  const data = await api.getVersionData(version).catch((err) => {
      console.error(`Failed to get version data: ${err.message}`)
      Deno.exit(1)
  })

  version = data.id

  const rootDir = `${Deno.cwd()}/minecraft`
  const instanceDir = `${rootDir}/instances/${flags.name ?? version}`

  if (!await exists(rootDir)) {
    await Deno.mkdir(rootDir)
  }
  if (!await exists(instanceDir)) {
    await Deno.mkdir(instanceDir, {recursive:true})
  }

  Deno.chdir(rootDir)

  console.log("Loading libraries...")
  for (const library of data.libraries) {
    util.writeOnLine(library.downloads.artifact.path)
    await api.downloadLibrary(library, "libraries")
  }

  console.log("\nDownloading assets...")
  const assets = await api.getAssetData(data.assetIndex.url)
  const numberOfAssets = Object.keys(assets.objects).length
  let i = 0
  for (const [name, asset] of Object.entries(assets.objects)) {
    i++
    util.writeOnLine(`${i}/${numberOfAssets} ${name}`)
    await api.downloadAsset(asset, "assets")
  }
  await util.saveFile(JSON.stringify(assets), `assets/indexes/${data.assetIndex.id}.json`)

  const clientPath = `${instanceDir}/client.jar`
  if (!await exists(clientPath)) {
    console.log("\nDownloading client...")
    await util.download(data.downloads.client.url, clientPath)
  }


  const classPath = `${clientPath}:${util.getLibraryPaths(data.libraries, `${rootDir}/libraries`)}`

  const cmdArgs = {
    java: ["-cp", classPath],
    game: ["--version", "", "--accessToken", "abc", "--assetsDir", "assets", "--assetIndex", data.assetIndex.id, "--gameDir", instanceDir]
  }

  if (flags.username) {
    cmdArgs.game.push("--username", flags.username)
  }
  if (Deno.build.os == "darwin") {
    cmdArgs.java.push("-XstartOnFirstThread")
  }
  console.log(`\nStarting Minecraft ${version}...`)
  new Deno.Command("java", { args: [...cmdArgs.java, data.mainClass, ...cmdArgs.game] }).spawn()
}


if (import.meta.main) {
  main(Deno.args)
}