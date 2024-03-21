import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { dirname } from "https://deno.land/std@0.219.1/path/dirname.ts";
import { Library } from "./types.ts";

export const writeOnLine = (s: string) => { Deno.stdout.write(new TextEncoder().encode("\x1b[1K\r" + s)) }

export function getLibraryPaths(libraries: Library[], rootDir: string) {
    const paths: string[] = []
    libraries.forEach((element) => {
      paths.push(`${rootDir}/${element.downloads.artifact.path}`)
  })
    return paths.join(":")
}

export async function download(url: string, dest: string) {
  const data = await (await fetch(url)).arrayBuffer()
  const dir = dirname(dest)
  if (!await exists(dir)) {
    await Deno.mkdir(dir, {recursive:true})
  }

  await Deno.writeFile(dest, new Uint8Array(data))
}

export async function saveFile(data: string, path: string) {
  const dir = dirname(path)
  if (!await exists(dir)) {
    await Deno.mkdir(dir, {recursive:true})
  }
  await Deno.writeFile(path, new TextEncoder().encode(data))
}