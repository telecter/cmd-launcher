import { type VersionData, type AssetData } from "./types.ts"
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
    string: ["version", "username"],
    boolean: ["help", "list-releases", "list-snapshots", "update"],
  })
  if (flags.help) {
    console.log(`
Command Line Minecraft Launcher v0.3.1-alpha
Usage: minecraft-launcher [options]
--version         Version of Minecraft to launch
--username        Username, defaults to random
--list-snapshots  List all snapshots
--list-releases   List all releases
--update          Update the launcher
--help            Display this help and exit
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
  if (flags["list-releases"]) {
    (await api.filterVersions("release")).forEach((element) => console.log(element.id))
    Deno.exit()
  }
  else if (flags["list-snapshots"]) {
    (await api.filterVersions("snapshot")).forEach((element) => console.log(element.id))
    Deno.exit()
  }
  console.log(`Getting version data for ${version ?? "latest"}`)
  const data: VersionData = await api.getVersionData(version).catch((err) => {
      console.error(`Failed to get version data: ${err.message}`)
      Deno.exit(1)
  })

  version = data.id

  if (!await exists("minecraft")) {
    await Deno.mkdir("minecraft")
  }

  Deno.chdir("minecraft")

  if (!await exists(version)) {
    await Deno.mkdir(version, {recursive:true})
  }


  if (!await exists("libraries")) {
    console.log("Downloading libraries...")
    for (const library of data.libraries) {
      util.writeOnLine(library.downloads.artifact.path)
      await api.downloadLibrary(library, Deno.cwd())
    }
  }

  Deno.chdir(version)

  if (!await exists("client.jar")) {
    console.log("\nDownloading client...")
    await util.download(data.downloads.client.url, "client.jar")
  }

  if (!await exists("assets")) {
    console.log("\nDownloading assets...")
    const assets: AssetData = await (await fetch(data.assetIndex.url)).json()
    const numberOfAssets = Object.keys(assets.objects).length
    let i = 0;
    for (const [name, asset] of Object.entries(assets.objects)) {
      util.writeOnLine(`${i}/${numberOfAssets} ${name}`)
      await api.downloadAsset(asset, Deno.cwd())
      i++
    }
    await api.downloadAssetData(data.assetIndex.url, data.assetIndex.id, Deno.cwd())
  }

  const cmdArgs = {
    java: ["-cp", `client.jar:${util.getLibraryPaths(data.libraries, "../libraries").join(":")}`, "-XstartOnFirstThread", data.mainClass],
    game: ["--version", "", "--accessToken", "abc", "--assetsDir", "assets", "--assetIndex", data.assetIndex.id, "--gameDir", Deno.cwd()]
  }
  if (flags.username) {
    cmdArgs.game.push("--username", flags.username)
  }
  console.log(`\nStarting Minecraft ${version}...`)
  new Deno.Command("java", { args: cmdArgs.java.concat(cmdArgs.game) }).spawn()
}


if (import.meta.main) {
  main(Deno.args)
}