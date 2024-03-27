import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { Asset, Library, VersionManifest } from "./types.ts";
import { download, saveFile } from "./util.ts";
import { AssetIndexData, AssetIndex, VersionData } from "./types.ts";

/** Downloads the Minecraft version manifest, containing the URL and ID for each version. */
export async function getVersionManifest(): Promise<VersionManifest> {
  return await (
    await fetch("https://launchermeta.mojang.com/mc/game/version_manifest.json")
  ).json();
}

/**
Get the data for a specific version. This includes all the information needed to launch the game.
If the version is null, the latest version of the game is used.
*/
export async function getVersionData(
  version: string | null,
): Promise<VersionData> {
  const data = await getVersionManifest();
  if (!version) {
    version = data.latest.release;
  }
  const release = data.versions.find((element) => element.id == version);
  if (!release) {
    throw Error("Invalid version");
  }
  return await (await fetch(release.url)).json();
}

/** Download the asset data from the specified URL, save it to the `assets` directory, and return the JSON file. Only downloads if the file does not already exist. */
export async function downloadAssetData(
  assetIndex: AssetIndexData,
  rootDir: string,
): Promise<AssetIndex> {
  const data = await (await fetch(assetIndex.url)).json();
  const path = `${rootDir}/assets/indexes/${assetIndex.id}.json`;
  if (!(await exists(path))) {
    await saveFile(
      JSON.stringify(data),
      `${rootDir}/assets/indexes/${assetIndex.id}.json`,
    );
  }
  return data;
}

/** Download the specified game library to the `libraries` directory. Only downloads if the file does not already exist. */
export async function downloadLibrary(library: Library, rootDir: string) {
  const artifact = library.downloads.artifact;
  const path = `${rootDir}/libraries/${artifact.path}`;
  if (!(await exists(path))) {
    await download(artifact.url, path);
  }
}

/** Download the specified asset to the `assets` directory. Only downloads if the file does not already exist. */
export async function downloadAsset(asset: Asset, rootDir: string) {
  const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`;
  const path = `${rootDir}/assets/objects/${objectPath}`;
  if (!(await exists(path))) {
    await download(
      `https://resources.download.minecraft.net/${objectPath}`,
      path,
    );
  }
}
