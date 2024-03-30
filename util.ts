import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { dirname } from "https://deno.land/std@0.219.1/path/dirname.ts";

export async function download(url: string, dest: string) {
  const data = await (await fetch(url)).arrayBuffer();
  const dir = dirname(dest);
  if (!(await exists(dir))) {
    await Deno.mkdir(dir, { recursive: true });
  }

  await Deno.writeFile(dest, new Uint8Array(data));
}

export async function saveFile(data: string, path: string) {
  const dir = dirname(path);
  if (!(await exists(dir))) {
    await Deno.mkdir(dir, { recursive: true });
  }
  await Deno.writeFile(path, new TextEncoder().encode(data));
}

export function log(s: string, level: "INFO" | "ERROR" = "INFO") {
  const output = `[%c${level}%c] ${s}`;
  if (level == "INFO") {
    console.log(output, "color: blue", "");
  } else if (level == "ERROR") {
    console.error(output, "color: red", "");
  }
}
