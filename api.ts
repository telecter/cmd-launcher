import { dirname } from "https://deno.land/std@0.219.1/path/dirname.ts";
import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { type Asset, type Library, type VersionManifest } from "./types.ts";

export async function download(url: string, dest: string) {
    const data = await (await fetch(url)).arrayBuffer()
    const dir = dirname(dest)
    if (!await exists(dir)) {
      await Deno.mkdir(dir, {recursive:true})
    }
    await Deno.writeFile(dest, new Uint8Array(data))
    return dest
}

export async function getVersionManifest() {
  return <VersionManifest>(await (await fetch("https://launchermeta.mojang.com/mc/game/version_manifest.json")).json())
}

export async function filterVersions(filter: "release"|"snapshot") {
  return (await getVersionManifest()).versions.filter((element) => element.type == filter)
}


export async function getVersionData(version: string|null) {
    const data = await getVersionManifest()
    if (!version) {
      version = data.latest.release
    }
    const release = data.versions.find((element) => element.id == version)
    if (!release) {
      throw Error("Invalid version")
    }
    return (await fetch(release.url)).json()
}

export async function downloadLibrary(library: Library) {
    const artifact = library.downloads.artifact
    await download(artifact.url, `libraries/${artifact.path}`)
}
export async function downloadAsset(asset: Asset) {
      const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`
      const path = `assets/objects/${objectPath}`
      await download(`https://resources.download.minecraft.net/${objectPath}`, path)
  }
