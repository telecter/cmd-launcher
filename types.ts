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
export type Version = {
  id: string;
  type: "release" | "snapshot";
  url: string;
  time: string;
  releaseTime: string;
};

export type VersionManifest = {
  latest: {
    snapshot: string;
    release: string;
  };
  versions: Version[];
};

export type VersionMeta = {
  assetIndex: AssetIndexMeta;
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

export type AssetIndexMeta = {
  id: string;
  sha1: string;
  size: number;
  totalSize: number;
  url: string;
};

export type AssetIndex = {
  objects: { [name: string]: Asset };
};

export type VersionOptions = {
  auth?: AuthData;
  username?: string;
  rootDir: string;
  instanceDir: string;
  loader?: string;
  jvmPath: string;
  cache: boolean;
};

export type LaunchArgs = {
  mainClass: string;
  assetId: string;
  client: string;
  libraries: string[];
};

export type AuthData = {
  username: string;
  uuid: string;
  token: string;
  refresh: string;
};
