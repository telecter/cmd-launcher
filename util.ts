import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import { walk } from "https://deno.land/std@0.219.1/fs/walk.ts";

export async function initDirectory(version: string) {
    const versionPath = `minecraft/${version}`
    if (!await exists(versionPath)) {
        await Deno.mkdir(versionPath, {recursive:true})
    }
    return versionPath
}
export async function findLibraries() {
    const libPaths = []
    for await (const walkEntry of walk("libraries")) {
      if (walkEntry.isFile) {
        libPaths.push(walkEntry.path)
      }
    }
    return libPaths
}