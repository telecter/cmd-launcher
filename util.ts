import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { dirname } from "https://deno.land/std@0.219.1/path/dirname.ts";
import { ensureDir } from "https://deno.land/std@0.221.0/fs/mod.ts";

export async function saveTextFile(path: string, data: string) {
  await ensureDir(dirname(path));
  await Deno.writeTextFile(path, data);
}

export async function fetchJSONData(url: string) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`Request to ${url} failed with status code ${res.status}`);
  }
  const data = await res.json().catch((_err) => {
    throw new Error(`Failed to parse JSON response from ${url}`);
  });
  return data;
}

export function getPathFromMaven(mavenPath: string) {
  const dir = mavenPath.replace(".", "/").split(":");
  const basename = `${dir[1]}-${dir[2]}.jar`;
  const path = `${dir.join("/")}/${basename}`.replace("ow2.asm", "ow2/asm");
  return path;
}

export async function readJSONIfExists(path: string) {
  if (await exists(path)) {
    return JSON.parse(await Deno.readTextFile(path));
  }
  return null;
}
