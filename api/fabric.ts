import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { download } from "../util.ts";

function getPathFromMaven(mavenPath: string) {
  const dir = mavenPath.replace(".", "/").split(":");
  const basename = `${dir[1]}-${dir[2]}.jar`;
  const path = `${dir.join("/")}/${basename}`.replace("ow2.asm", "ow2/asm");
  return path;
}

export async function fetchFabricLibraries(
  libraries: { name: string; url: string }[],
  rootDir: string,
) {
  const paths = [];
  for (const library of libraries) {
    const path = getPathFromMaven(library.name);
    const fsPath = `${rootDir}/libraries/${path}`;
    if (!(await exists(fsPath))) {
      await download(`${library.url}/${path}`, fsPath);
    }
    paths.push(fsPath);
  }
  return paths;
}

export async function getFabricMeta(gameVersion: string) {
  const versionList = await (
    await fetch(`https://meta.fabricmc.net/v2/versions/loader/${gameVersion}`)
  ).json();

  return (
    await fetch(
      `https://meta.fabricmc.net/v2/versions/loader/${gameVersion}/${versionList[0].loader.version}/profile/json`,
    )
  ).json();
}

export async function getQuiltMeta(gameVersion: string) {
  return (
    await fetch(
      `https://meta.quiltmc.org/v3/versions/loader/${gameVersion}/0.24.0/profile/json`,
    )
  ).json();
}
