import { VersionOptions } from "./types.ts";
import * as api from "./api/game.ts";
import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { saveFile, getPathFromMaven } from "./util.ts";
import { getFabricMeta, getQuiltMeta } from "./api/fabric.ts";
import { LaunchData } from "./types.ts";
import { dirname } from "https://deno.land/std@0.219.1/path/windows/dirname.ts";
import { ensureDir } from "https://deno.land/std@0.221.0/fs/ensure_dir.ts";

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
  const meta = await api.getVersionData(version);

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
  if (options.fabric || options.quilt) {
    const fabricMeta = options.quilt
      ? await getQuiltMeta(version)
      : await getFabricMeta(version);
    mainClass = fabricMeta.mainClass;
    for (const library of fabricMeta.libraries) {
      const path = getPathFromMaven(library.name);
      const fsPath = `${options.rootDir}/libraries/${path}`;
      if (!(await exists(fsPath))) {
        const url = `${library.url}/${path}`;
        await download(url, fsPath);
      }
      libraries.push(fsPath);
    }
  }

  const assets = await api.getAssetData(meta);
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
    await saveFile(JSON.stringify(assets), assetIndexPath);
  }
  const clientUrl = meta.downloads.client.url;
  const clientPath = `${options.instanceDir}/client.jar`;
  if (!(await exists(clientPath))) {
    await download(clientUrl, clientPath);
  }
  return <LaunchData>{
    mainClass: mainClass,
    assetId: meta.assetIndex.id,
    client: clientPath,
    libraries: libraries,
  };
}

export function run(meta: LaunchData, options: VersionOptions) {
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
    options.auth?.username ?? options.offlineUsername ?? "",
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
