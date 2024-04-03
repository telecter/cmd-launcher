import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { dirname } from "https://deno.land/std@0.219.1/path/dirname.ts";

export async function saveFile(data: string, path: string) {
  const dir = dirname(path);
  if (!(await exists(dir))) {
    await Deno.mkdir(dir, { recursive: true });
  }
  await Deno.writeFile(path, new TextEncoder().encode(data));
}

export function getPathFromMaven(mavenPath: string) {
  const dir = mavenPath.replace(".", "/").split(":");
  const basename = `${dir[1]}-${dir[2]}.jar`;
  const path = `${dir.join("/")}/${basename}`.replace("ow2.asm", "ow2/asm");
  return path;
}
