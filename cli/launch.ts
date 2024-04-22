import { registerDownloadListener, saveTextFile } from "../util.ts";
import * as launcher from "../launcher.ts";
import { VersionOptions } from "../launcher.ts";
import { ensureDir, exists } from "https://deno.land/std@0.221.0/fs/mod.ts";
import { getAuthData } from "../api/auth.ts";
import { getDirs } from "./main.ts";
import { Command } from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/command.ts";

async function launch(
  flags: { java?: string; server?: string; username?: string; dir?: string },
  ...args: string[]
) {
  registerDownloadListener((url) => console.log(`Downloading ${url}`));
  const [version, loader] = args[0].split(":");

  if (loader && !launcher.MOD_LOADERS.includes(loader)) {
    throw Error("Invalid mod loader.");
  }
  const [rootDir, instanceDir] = getDirs(version, flags.dir);

  await ensureDir(rootDir);

  const authCache = `${rootDir}/accounts.json`;

  const options: VersionOptions = {
    rootDir: rootDir,
    instanceDir: instanceDir,
    jvmPath: flags.java ?? "java",
    cache: true,
    server: flags.server,
  };
  if (loader) {
    options.loader = loader;
  }

  if (!flags.username) {
    try {
      if (await exists(authCache)) {
        const refreshToken = JSON.parse(
          await Deno.readTextFile(authCache),
        ).refresh;
        options.auth = await getAuthData(refreshToken);
      } else {
        options.auth = await getAuthData();
      }
      await saveTextFile(
        authCache,
        JSON.stringify({
          refresh: options.auth?.refresh,
        }),
      );
    } catch (err) {
      console.error(`
        Authentication failed: ${err.message}. Using offline mode.`);
    }
  } else {
    options.username = flags.username;
  }

  const instance = await launcher.install(version, options);
  console.log(`Starting Minecraft ${version}...`);
  launcher.run(instance);
}

export const launchCommand = new Command()
  .description("Launch the game with the specified options.")
  .arguments("<version:string>")
  .option("-j, --java <path:string>", "Use the specified Java executable.")
  .option("-u, --username <username:string>", "Set offline username")
  .option("-s, --server <server:string>", "Join specified server on launch")
  .action(launch);
