import { VersionData } from "./types.ts";
import { VersionOptions } from "./types.ts";
import * as api from "./api/game.ts";
import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { download, saveFile, getPathFromMaven } from "./util.ts";
import { getFabricMeta, getQuiltMeta } from "./api/fabric.ts";

export interface Version {
  releaseVersion: string;
  options: VersionOptions;
  libraries: string[];
  meta: VersionData;
  fabricMeta?: any;
  client: string;
  mainClass: string;
}

export class Version {
  constructor(releaseVersion: string, options: VersionOptions) {
    this.releaseVersion = releaseVersion;
    this.options = options;
    this.client = `${this.options.instanceDir}/client.jar`;
    this.libraries = [];
  }
  /** Initialize the version data. Things will not work correctly if you don't run this function before installing or running the version. */
  async init() {
    this.meta = await api.getVersionData(this.releaseVersion);
    if (this.options.fabric || this.options.quilt) {
      this.fabricMeta = this.options.fabric
        ? await getFabricMeta(this.releaseVersion)
        : await getQuiltMeta(this.releaseVersion);
      this.mainClass = this.fabricMeta.mainClass;
    } else {
      this.mainClass = this.meta.mainClass;
    }
  }
  async install(
    downloadListener: (url: string) => void = (_url: string) => {},
  ) {
    for (const library of this.meta.libraries) {
      const artifact = library.downloads.artifact;
      const path = `${this.options.rootDir}/libraries/${artifact.path}`;
      if (!(await exists(path))) {
        await download(artifact.url, path);
        downloadListener(artifact.url);
      }
      this.libraries.push(path);
    }
    if (this.fabricMeta) {
      for (const library of this.fabricMeta.libraries) {
        const path = getPathFromMaven(library.name);
        const fsPath = `${this.options.rootDir}/libraries/${path}`;
        if (!(await exists(fsPath))) {
          const url = `${library.url}/${path}`;
          await download(url, fsPath);
          downloadListener(url);
        }
        this.libraries.push(fsPath);
      }
    }

    const assets = await api.getAssetData(this.meta);
    for (const asset of Object.values(assets.objects)) {
      const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`;
      const path = `${this.options.rootDir}/assets/objects/${objectPath}`;
      if (!(await exists(path))) {
        const url = `https://resources.download.minecraft.net/${objectPath}`;
        await download(url, path);
        downloadListener(url);
      }
    }
    const assetIndexPath = `${this.options.rootDir}/assets/indexes/${this.meta.assetIndex.id}.json`;
    if (!(await exists(assetIndexPath))) {
      await saveFile(JSON.stringify(assets), assetIndexPath);
    }
    const clientUrl = this.meta.downloads.client.url;
    if (!(await exists(this.client))) {
      await download(clientUrl, `${this.options.instanceDir}/client.jar`);
      downloadListener(clientUrl);
    }
  }
  run() {
    const classPath = [this.client, ...this.libraries];
    const jvmArgs = ["-cp", classPath.join(":")];
    if (Deno.build.os === "darwin") {
      jvmArgs.push("-XstartOnFirstThread");
    }
    const gameArgs = [
      "--version",
      "",
      "--accessToken",
      this.options.accessToken,
      "--uuid",
      this.options.uuid,
      "--username",
      this.options.username,
      "--assetsDir",
      `${this.options.rootDir}/assets`,
      "--assetIndex",
      this.meta.assetIndex.id,
    ];
    Deno.chdir(this.options.instanceDir);
    new Deno.Command("java", {
      args: [...jvmArgs, this.mainClass, ...gameArgs],
    }).spawn();
  }
}
