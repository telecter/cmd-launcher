import { type Asset, type Library, type VersionManifest } from "./types.ts";
import { download } from "./util.ts";

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

export async function downloadAssetData(url: string, id: string, rootDir: string) {
  await download(url, `${rootDir}/assets/indexes/${id}.json`)
}

export async function downloadLibrary(library: Library, rootDir: string) {
    const artifact = library.downloads.artifact
    await download(artifact.url, `${rootDir}/libraries/${artifact.path}`)
}
export async function downloadAsset(asset: Asset, rootDir: string) {
      const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`
      const path = `${rootDir}/assets/objects/${objectPath}`
      await download(`https://resources.download.minecraft.net/${objectPath}`, path)
}
