import { fetchJSONData } from "../util.ts";

export type VersionManifest = {
  latest: {
    snapshot: string;
    release: string;
  };
  versions: {
    id: string;
    type: "release" | "snapshot";
    url: string;
    time: string;
    releaseTime: string;
  }[];
};
export type VersionMeta = {
  assetIndex: {
    id: string;
    sha1: string;
    size: number;
    totalSize: number;
    url: string;
  };
  assets: string;
  downloads: {
    client: {
      sha1: string;
      size: number;
      url: string;
    };
  };
  id: string;
  javaVersion: {
    component: string;
    majorVersion: number;
  };
  libraries: Library[];
  mainClass: string;
};

export type Library = {
  downloads: {
    artifact: {
      path: string;
      sha1: string;
      size: number;
      url: string;
    };
  };
  url?: string;
  name: string;
};
export type Asset = {
  hash: string;
  size: number;
};
export type AssetIndex = {
  objects: { [name: string]: Asset };
};

/** Downloads the Minecraft version manifest, containing the URL and ID for each version. */
export async function getVersionManifest() {
  const data = await fetchJSONData(
    "https://launchermeta.mojang.com/mc/game/version_manifest.json",
  );
  return data as VersionManifest;
}

/**
Get the data for a specific version. This includes all the information needed to launch the game.
*/
export async function getVersionMeta(version: string) {
  const manifest = await getVersionManifest();
  const release = manifest.versions.find((element) => element.id == version);
  if (!release) {
    throw Error("Invalid version");
  }
  const meta = await fetchJSONData(release.url);
  return meta as VersionMeta;
}

/** Using the version metadata, get the game asset data. */
export async function getAssetData(meta: VersionMeta) {
  const url = meta.assetIndex.url;
  const data = await fetchJSONData(url);
  return data as AssetIndex;
}
