import { type VersionData, type AssetData } from "./types.ts"
import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.207.0/cli/parse_args.ts";
import * as api from "./api.ts"
import * as util from "./util.ts"
const writeOnLine = (s: string) => { Deno.stdout.write(new TextEncoder().encode("\x1b[1K\r" + s)) }

async function update() {
  const tags = await (await fetch("https://api.github.com/repos/telectr/minecraft-launcher/tags")).json()
  const latestTag = tags[0].name
  console.log(latestTag)
  const data = await (await fetch(`https://github.com/telectr/minecraft-launcher/releases/download/${latestTag}/launcher-${Deno.build.target}`)).arrayBuffer()
  if (Deno.execPath().includes("deno")) {
    throw Error("Cannot update non-executable")
  }
  await Deno.writeFile(Deno.execPath(), new Uint8Array(data))
  Deno.exit()
}

async function getArgs(args: string[]) {
  const flags = parseArgs(args, {
    string: ["version", "username"],
    boolean: ["help", "list-releases", "list-snapshots", "update"]
  })
  if (flags.help) {
    console.log(`
Command Line Minecraft Launcher
Usage: minecraft-launcher [options]
--version     Version of Minecraft to launch
--username    Username, defaults to random
--update      Update the launcher
--help        Display this help and exit
    `)
    Deno.exit()
  }
  else if (flags["list-releases"] || flags["list-snapshots"]) {
    let versionType: string
    if (flags["list-releases"]) {
      versionType = "release"
    }
    else if (flags["list-snapshots"]) {
      versionType = "snapshot"
    }
    api.getVersionManifest().then((data) => {
      data.versions.filter((element) => element.type == versionType).forEach((element) => console.log(element.id))
    }).finally(() => Deno.exit())
  }
  else if (flags.update) {
    await update()
    Deno.exit(0)
  }

  return flags
}
async function main(args: string[]) {
  console.log("a")
  const flags = await getArgs(args)
  let version = null
  if (flags.version) {
    version = flags.version
  }

  console.log(`Getting version data for ${version ?? "latest"}`)
  const data: VersionData = await api.getVersionData(version).catch((err) => {
      console.error(`Failed to get version data: ${err.message}`)
      Deno.exit(1)
  })
  version = data.id
  Deno.chdir(await util.initDirectory(version))

  if (!await exists("client.jar")) {
    console.log("Downloading client...")
    await api.download(data.downloads.client.url, "client.jar")
  }
  if (!await exists("libraries")) {
    console.log("Downloading libraries...")
    for (const library of data.libraries) {
      writeOnLine(library.downloads.artifact.path)
      await api.downloadLibrary(library)
    }
  }
  if (!await exists("assets")) {
    console.log("\nDownloading assets...")
    const assets: AssetData = await (await fetch(data.assetIndex.url)).json()
    const numberOfAssets = Object.keys(assets.objects).length
    let i = 0;
    for (const [name, asset] of Object.entries(assets.objects)) {
      writeOnLine(`${i}/${numberOfAssets} ${name}`)
      await api.downloadAsset(asset)
      i++
    }
    await api.download(data.assetIndex.url, `assets/indexes/${data.assetIndex.id}.json`)
  }

  const cmdArgs = {
    java: ["-cp", `client.jar:${(await util.findLibraries()).join(":")}`, "-XstartOnFirstThread", data.mainClass],
    game: ["--version", "", "--accessToken", "abc", "--assetsDir", "assets", "--assetIndex", data.assetIndex.id]
  }
  console.log(`\nStarting Minecraft ${version}...`)
  new Deno.Command("java", { args: cmdArgs.java.concat(cmdArgs.game) }).spawn()
}


if (import.meta.main) {
  main(Deno.args)
}