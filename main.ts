import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.220.1/cli/mod.ts"
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


async function main(args: string[]) {
  const flags = parseArgs(args, { 
    string: ["version", "username"],
    boolean: ["help"],
    alias: {
      "version": "v",
      "username": "u",
      "help": "help"
    },
    unknown: (arg) => { 
      console.log(`Invalid argument: ${arg}`)
      Deno.exit(1)
     }
   })
  let version = flags.version ?? null

  if (flags.help) {
    console.log(`
    usage: minecraft-launcher [...options]
    Options:
      -h, --help      Display this help and exit
      -v, --version   Set version of game to launcher
      -u, --username  Set username, defaults to random
    `)
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


  const classPath = [clientPath, ...data.libraries.map((element) => `libraries/${element.downloads.artifact.path}`)].join(":")

  const javaArgs = ["-cp", classPath]
  const gameArgs = ["--version", "", "--accessToken", "abc", "--assetsDir", "assets", "--assetIndex", data.assetIndex.id, "--gameDir", instanceDir]
  if (Deno.build.os == "darwin") {
    javaArgs.push("-XstartOnFirstThread")
  }
  if (flags.username) {
    gameArgs.push("--username", flags.username)
  }
  console.log(`\nStarting Minecraft ${version}...`)
  new Deno.Command("java", {
    args: [...javaArgs, data.mainClass, ...gameArgs]
  }).spawn()
}


if (import.meta.main) {
  main(Deno.args)
}