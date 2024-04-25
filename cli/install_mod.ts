import { ensureDir } from "https://deno.land/std@0.221.0/fs/mod.ts";
import { getProjectVersion } from "../api/modrinth.ts";
import { getDirs } from "./main.ts";
import { download } from "../util.ts";
import { Command } from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/command.ts";

async function installMods(flags: { dir?: string }, ...args: string[]) {
  const gameVersion = args[1];
  const id = args[0];

  const version = await getProjectVersion(id, gameVersion);

  const [_rootDir, instanceDir] = getDirs(gameVersion, flags.dir);
  const modsDir = `${instanceDir}/mods`;
  await ensureDir(modsDir);

  const file = version.files[0];
  await download(file.url, `${modsDir}/${file.filename}`);
  console.log(`Successfully downloaded jar file: ${file.filename}`);
}
export const installModsCommand = new Command()
  .description("Install and manage mods.")
  .command("install")
  .arguments("<mod:string> <version:string>")
  // @ts-ignore Does not think global option exists.
  .action(installMods);
