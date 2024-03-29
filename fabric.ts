import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { download } from "./util.ts";

function getPathFromMaven(mavenPath: string) {
  const dir = mavenPath.replace(".", "/").split(":");
  const basename = `${dir[1]}-${dir[2]}.jar`;
  const path = `${dir.join("/")}/${basename}`.replace("ow2.asm", "ow2/asm");
  return path;
}

export async function fetchFabricLibrary(library: any, rootDir: string) {
  const path = getPathFromMaven(library.name);
  const fsPath = `${rootDir}/libraries/${path}`;
  if (!(await exists(fsPath))) {
    await download(`https://maven.fabricmc.net/${path}`, fsPath);
  }
  return fsPath;
}

export async function getFabricMeta(gameVersion: string) {
  const versionList = await (
    await fetch("https://meta.fabricmc.net/v2/versions/loader/1.20.4")
  ).json();

  return (
    await fetch(
      `https://meta.fabricmc.net/v2/versions/loader/${gameVersion}/${versionList[0].loader.version}/profile/json`,
    )
  ).json();
}
