import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { download, saveFile } from "../util.ts";
import { AssetIndex, VersionData, Library, VersionManifest } from "../types.ts";

let downloadListener = (_url: string) => {};
export function registerDownloadListener(callback: (url: string) => void) {
  downloadListener = callback;
}

/** Downloads the Minecraft version manifest, containing the URL and ID for each version. */
export async function getVersionManifest() {
  return <VersionManifest>(
    await (
      await fetch(
        "https://launchermeta.mojang.com/mc/game/version_manifest.json",
      )
    ).json()
  );
}

/**
Get the data for a specific version. This includes all the information needed to launch the game.
If the version is null, the latest version of the game is used.
*/
export async function getVersionData(version: string | null) {
  const data = await getVersionManifest();
  if (!version) {
    version = data.latest.release;
  }
  const release = data.versions.find((element) => element.id == version);
  if (!release) {
    throw Error("Invalid version");
  }
  return <VersionData>await (await fetch(release.url)).json();
}

/** Download the asset data, save it to the `assets` directory in the root directory, and return the JSON data. */
export async function getAndSaveAssetData(
  versionData: VersionData,
  rootDir: string,
) {
  const url = versionData.assetIndex.url;
  const data = await (await fetch(url)).json();
  const path = `${rootDir}/assets/indexes/${versionData.assetIndex.id}.json`;
  if (!(await exists(path))) {
    await saveFile(JSON.stringify(data), path);
    downloadListener(url);
  }
  return <AssetIndex>data;
}

/** Fetch the specified game libraries to the `libraries` directory. Only downloads if the file does not already exist. */
export async function fetchLibraries(libraries: Library[], rootDir: string) {
  const paths: string[] = [];
  for (const library of libraries) {
    const artifact = library.downloads.artifact;
    const path = `${rootDir}/libraries/${artifact.path}`;
    if (!(await exists(path))) {
      await download(artifact.url, path);
      downloadListener(artifact.url);
    }
    paths.push(path);
  }
  return paths;
}

/** Fetch the specified assets to the `assets` directory. Only downloads if the file does not already exist. */
export async function fetchAssets(assets: AssetIndex, rootDir: string) {
  for (const asset of Object.values(assets.objects)) {
    const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`;
    const path = `${rootDir}/assets/objects/${objectPath}`;
    if (!(await exists(path))) {
      const url = `https://resources.download.minecraft.net/${objectPath}`;
      await download(url, path);
      downloadListener(url);
    }
  }
}

export async function fetchClient(versionData: VersionData, rootDir: string) {
  const clientPath = `${rootDir}/client.jar`;
  if (!(await exists(clientPath))) {
    const url = versionData.downloads.client.url;
    await download(versionData.downloads.client.url, `${rootDir}/client.jar`);
    downloadListener(url);
  }
  return clientPath;
}
