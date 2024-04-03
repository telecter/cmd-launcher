export type Library = {
  downloads: {
    classifiers?: {
      "natives-linux": Artifact;
      "natives-osx": Artifact;
      "natives-windows": Artifact;
    };
    artifact: Artifact;
  };
  name: string;
};
export type Artifact = {
  path: string;
  sha1: string;
  size: number;
  url: string;
};

export type Asset = {
  hash: string;
  size: number;
};
export type VersionIdentifier = {
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
  versions: VersionIdentifier[];
};

export type VersionData = {
  assetIndex: AssetIndexData;
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

export type AssetIndexData = {
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
  accessToken: string;
  uuid: string;
  username: string;
  rootDir: string;
  instanceDir: string;
  fabric?: boolean;
  quilt?: boolean;
};
