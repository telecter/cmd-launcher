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
import { saveTextFile } from "./util.ts";

launcher.registerDownloadListener((url) => console.log(`Downloading ${url}`));

async function launch(
  flags: { java?: string; server?: string; username?: string; dir?: string },
  ...args: string[]
) {
  const [version, loader] = args[0].split(":");

  if (loader && !launcher.MOD_LOADERS.includes(loader)) {
    console.error("Invalid mod loader");
    Deno.exit(1);
  }

  const rootDirString = flags.dir ?? `${Deno.env.get("HOME")}/.minecraft`;
  const rootDir = isAbsolute(rootDirString)
    ? rootDirString
    : `${Deno.cwd()}/${rootDirString}`;
  const instanceDir = `${rootDir}/instances/${version}`;

  await ensureDir(rootDir);

  const accountDataFile = `${rootDir}/accounts.json`;

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
      if (await exists(accountDataFile)) {
        const refreshToken = JSON.parse(
          await Deno.readTextFile(accountDataFile),
        ).refresh;
        options.auth = await getAuthData(refreshToken);
      } else {
        options.auth = await getAuthData();
      }
      await saveTextFile(
        accountDataFile,
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
    .command("launch", "Launch the game with the specified options.")
    .arguments("<version:string>")
    .option("-j, --java <path:string>", "Use the specified Java executable.")
    .option("-u, --username <username:string>", "Set offline username")
    .option("-s, --server <server:string>", "Join specified server on launch")
    .option("-d, --dir <path:string>", "Set the root directory for the game.")
    .action(launch)

    .command(
      "upgrade",
      new UpgradeCommand({
        provider: new GithubProvider({ repository: "telectr/cmd-launcher" }),
        main: "cli.ts",
        args: ["-A"],
      }),
    )

    .parse(Deno.args);
}
