import {
  VersionOptions,
  LaunchArgs,
  VersionMeta,
  AssetIndex,
} from "./types.ts";
import { getVersionMeta, getAssetData } from "./api/game.ts";
import { saveTextFile, getPathFromMaven, readJSONIfExists } from "./util.ts";
import { getFabricMeta, getQuiltMeta } from "./api/fabric.ts";
import { dirname } from "https://deno.land/std@0.221.0/path/dirname.ts";
import { ensureDir, exists } from "https://deno.land/std@0.221.0/fs/mod.ts";

export const MOD_LOADERS = ["fabric", "quilt"];

let downloadListener = (_url: string) => {};
export function registerDownloadListener(listener: (url: string) => void) {
  downloadListener = listener;
}

async function download(url: string, dest: string) {
  const data = await (await fetch(url)).arrayBuffer();
  downloadListener(url);
  await ensureDir(dirname(dest));
  await Deno.writeFile(dest, new Uint8Array(data));
  return dest;
}

export async function installVersion(version: string, options: VersionOptions) {
  const libraries = [];

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

  for (const library of meta.libraries) {
    const artifact = library.downloads.artifact;
    const path = `${options.rootDir}/libraries/${artifact.path}`;
    if (!(await exists(path))) {
      await download(artifact.url, path);
    }
    libraries.push(path);
  }
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

    for (const library of loaderMeta.libraries) {
      const path = getPathFromMaven(library.name);
      const fsPath = `${options.rootDir}/libraries/${path}`;
      if (!(await exists(fsPath))) {
        const url = `${library.url}/${path}`;
        await download(url, fsPath);
      }
      libraries.push(fsPath);
    }
  }

  const assetMetaCache = `${cachesDir}/assets/${meta.assetIndex.id}`;
  let assets: AssetIndex = await readJSONIfExists(assetMetaCache);
  if (!assets) {
    assets = await getAssetData(meta);
    await saveTextFile(assetMetaCache, JSON.stringify(assets));
  }

  for (const asset of Object.values(assets.objects)) {
    const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`;
    const path = `${options.rootDir}/assets/objects/${objectPath}`;
    if (!(await exists(path))) {
      const url = `https://resources.download.minecraft.net/${objectPath}`;
      await download(url, path);
    }
  }
  const assetIndexPath = `${options.rootDir}/assets/indexes/${meta.assetIndex.id}.json`;
  if (!(await exists(assetIndexPath))) {
    await saveTextFile(assetIndexPath, JSON.stringify(assets));
  }
  const clientUrl = meta.downloads.client.url;
  const clientPath = `${options.instanceDir}/${version}.jar`;
  if (!(await exists(clientPath))) {
    await download(clientUrl, clientPath);
  }
  return <LaunchArgs>{
    mainClass: mainClass,
    assetId: meta.assetIndex.id,
    client: clientPath,
    libraries: libraries,
  };
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
    options.auth?.token ?? "a",
    "--uuid",
    options.auth?.uuid ?? crypto.randomUUID(),
    "--username",
    options.auth?.username ??
      options.offlineUsername ??
      `Player${Math.floor(Math.random() * 100)}`,
    "--assetsDir",
    `${options.rootDir}/assets`,
    "--assetIndex",
    meta.assetId,
  ];
  Deno.chdir(options.instanceDir);
  new Deno.Command(options.jvmPath, {
    args: [...jvmArgs, meta.mainClass, ...gameArgs],
  }).spawn();
}
