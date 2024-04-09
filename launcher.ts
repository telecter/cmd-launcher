import { getVersionMeta, getAssetData } from "./api/game.ts";
import { saveTextFile, getPathFromMaven, readJSONIfExists } from "./util.ts";
import { getFabricMeta, getQuiltMeta } from "./api/fabric.ts";
import { dirname } from "https://deno.land/std@0.221.0/path/dirname.ts";
import { ensureDir, exists } from "https://deno.land/std@0.221.0/fs/mod.ts";
import {
  VersionOptions,
  LaunchArgs,
  VersionMeta,
  AssetIndex,
  Library,
} from "./types.ts";

export const MOD_LOADERS = ["fabric", "quilt"];

let downloadListener = (_url: string) => {};
export function registerDownloadListener(listener: (url: string) => void) {
  downloadListener = listener;
}

async function download(url: string, dest: string, overwrite: boolean = false) {
  if (!overwrite && (await exists(dest))) {
    return dest;
  }
  const data = await (await fetch(url)).arrayBuffer();
  downloadListener(url);
  await ensureDir(dirname(dest));
  await Deno.writeFile(dest, new Uint8Array(data));
  return dest;
}

/** Ensure, and if needed install, assets from the given version metadata. */
export async function installAssets(meta: VersionMeta, dir: string) {
  const cache = `${dir}/assets/indexes/${meta.assetIndex.id}.json`;
  let index: AssetIndex = await readJSONIfExists(cache);
  if (!index) {
    index = await getAssetData(meta);
    await saveTextFile(cache, JSON.stringify(index));
  }
  for (const asset of Object.values(index.objects)) {
    const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`;
    const path = `${dir}/assets/objects/${objectPath}`;
    await download(
      `https://resources.download.minecraft.net/${objectPath}`,
      path,
    );
  }
}

/** Ensure, and if needed install, game libraries from a given list. Returns the paths of the installed libraries. */
export async function installLibraries(libraries: Library[], dir: string) {
  const paths: string[] = [];
  for (const library of libraries) {
    let url: string, destPath: string;
    if (Object.hasOwn(library, "downloads")) {
      const artifact = library.downloads.artifact;
      url = library.downloads.artifact.url;
      destPath = `${dir}/libraries/${artifact.path}`;
    } else {
      const path = getPathFromMaven(library.name);
      const fsPath = `${dir}/libraries/${path}`;
      url = library.url + path;
      destPath = fsPath;
    }
    await download(url, destPath);
    paths.push(destPath);
  }
  return paths;
}

/** High level function for installing a version (libraries, assets, client). */
export async function install(version: string, options: VersionOptions) {
  const cachesDir = `${options.rootDir}/caches`;
  const versionMetaCache = `${cachesDir}/versions/${version}.json`;

  let meta: VersionMeta = await readJSONIfExists(versionMetaCache);
  if (!meta || !options.cache) {
    meta = await getVersionMeta(version);
    await saveTextFile(versionMetaCache, JSON.stringify(meta));
  }

  await ensureDir(options.rootDir);
  await ensureDir(options.instanceDir);

  let mainClass = meta.mainClass;

  let libraries = await installLibraries(meta.libraries, options.rootDir);
  console.log(libraries);
  if (options.loader) {
    const cachePath = `${cachesDir}/${options.loader === "quilt" ? "quilt" : "fabric"}/${version}.json`;
    let loaderMeta = await readJSONIfExists(cachePath);

    if (!loaderMeta || !options.cache) {
      loaderMeta =
        options.loader === "quilt"
          ? await getQuiltMeta(version)
          : await getFabricMeta(version);
      await saveTextFile(cachePath, JSON.stringify(loaderMeta));
    }

    mainClass = loaderMeta.mainClass;

    libraries = [
      ...libraries,
      ...(await installLibraries(loaderMeta.libraries, options.rootDir)),
    ];
  }

  await installAssets(meta, options.rootDir);

  const clientPath = `${options.instanceDir}/${version}.jar`;
  await download(meta.downloads.client.url, clientPath);

  return {
    mainClass: mainClass,
    assetId: meta.assetIndex.id,
    client: clientPath,
    libraries: libraries,
  } as LaunchArgs;
}

export function run(meta: LaunchArgs, options: VersionOptions) {
  const classPath = [meta.client, ...meta.libraries];

  const jvmArgs = ["-cp", classPath.join(":")];
  if (Deno.build.os === "darwin") {
    jvmArgs.push("-XstartOnFirstThread");
  }
  const gameArgs = [
    "--version",
    "",
    "--accessToken",
    options.auth?.token ?? "Haha this is not a valid access token.",
    "--uuid",
    options.auth?.uuid ?? crypto.randomUUID(),
    "--assetsDir",
    `${options.rootDir}/assets`,
    "--assetIndex",
    meta.assetId,
  ];
  if (options.auth?.username || options.username) {
    gameArgs.push("--username", options.auth?.username ?? options.username!);
  }
  Deno.chdir(options.instanceDir);
  new Deno.Command(options.jvmPath, {
    args: [...jvmArgs, meta.mainClass, ...gameArgs],
  }).spawn();
}
