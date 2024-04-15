import { exists, ensureDir } from "https://deno.land/std@0.221.0/fs/mod.ts";
import { isAbsolute } from "https://deno.land/std@0.221.0/path/is_absolute.ts";
import {
  Command,
  GithubProvider,
  UpgradeCommand,
} from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/mod.ts";
import * as launcher from "./launcher.ts";
import { getAuthData } from "./api/auth.ts";
import { VersionOptions } from "./launcher.ts";
import { download, registerDownloadListener, saveTextFile } from "./util.ts";
import { getProjectVersion } from "./api/modrinth.ts";

function getDirs(version: string, dir?: string) {
  const rawRootDir = dir ?? `${Deno.env.get("HOME")}/.minecraft`;
  const rootDir = isAbsolute(rawRootDir)
    ? rawRootDir
    : `${Deno.cwd()}/${rawRootDir}`;
  return [rootDir, `${rootDir}/instances/${version}`];
}

async function installMods(flags: { dir?: string }, ...args: string[]) {
  const gameVersion = args[0];
  const id = args[1];

  const version = await getProjectVersion(id, gameVersion).catch((err) => {
    console.error(`Failed to get project version: ${err.message}`);
    Deno.exit(1);
  });
  const [_rootDir, instanceDir] = getDirs(gameVersion, flags.dir);
  const modsDir = `${instanceDir}/mods`;
  await ensureDir(modsDir);
  console.log(modsDir);
  const file = version.files[0];
  await download(file.url, `${modsDir}/${file.filename}`);
  console.log(`Successfully downloaded jar file: ${file.filename}`);
}

async function launch(
  flags: { java?: string; server?: string; username?: string; dir?: string },
  ...args: string[]
) {
  registerDownloadListener((url) => console.log(`Downloading ${url}`));
  const [version, loader] = args[0].split(":");

  if (loader && !launcher.MOD_LOADERS.includes(loader)) {
    console.error("Invalid mod loader");
    Deno.exit(1);
  }

  const [rootDir, instanceDir] = getDirs(version, flags.dir);

  await ensureDir(rootDir);

  const authCache = `${rootDir}/accounts.json`;

  const options: VersionOptions = {
    rootDir: rootDir,
    instanceDir: instanceDir,
    jvmPath: flags.java ?? "java",
    cache: true,
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
      console.error(
        `Authentication failed: ${err.message}. Using offline mode.`,
      );
    }
  } else {
    options.username = flags.username;
  }

  const instance = await launcher.install(version, options).catch((err) => {
    console.error(`An error occurred while installing: ${err.message}`);
    Deno.exit(1);
  });
  console.log(`Starting Minecraft ${version}...`);
  try {
    launcher.run(instance);
  } catch (err) {
    console.error(`An error occurred while running: ${err.message}`);
    Deno.exit(1);
  }
}

if (import.meta.main) {
  const VERSION = "0.6.1";
  await new Command()
    .name("cmd-launcher")
    .version(VERSION)
    .description("A minimal command line Minecraft launcher.")
    .globalOption(
      "-d, --dir <path:string>",
      "Set the root directory for the game.",
    )
    .command("launch", "Launch the game with the specified options.")
    .arguments("<version:string>")
    .option("-j, --java <path:string>", "Use the specified Java executable.")
    .option("-u, --username <username:string>", "Set offline username")
    .option("-s, --server <server:string>", "Join specified server on launch")
    .action(launch)

    .command(
      "upgrade",
      new UpgradeCommand({
        provider: new GithubProvider({ repository: "telectr/cmd-launcher" }),
        main: "cli.ts",
        args: ["-A"],
      }),
    )

    .command("mod", "Install mods.")
    .arguments("<version:string> <mod:string>")
    .action(installMods)

    .parse(Deno.args);
}
