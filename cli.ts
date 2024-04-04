import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { Command } from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/mod.ts";
import { ensureDir } from "https://deno.land/std@0.221.0/fs/ensure_dir.ts";
import { installVersion, run, registerDownloadListener } from "./launcher.ts";
import { getAuthData } from "./api/auth.ts";
import { VersionOptions } from "./types.ts";
import {
  GithubProvider,
  UpgradeCommand,
} from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/upgrade/mod.ts";

type LaunchCmdOptions = {
  fabric?: boolean;
  quilt?: boolean;
  java?: string;
  server?: string;
  username?: string;
};

registerDownloadListener((url) => console.log(`Downloading ${url}`));

async function launchGame(flags: LaunchCmdOptions, ...args: string[]) {
  const version = args[0];

  const rootDir = `${Deno.cwd()}/minecraft`;
  await ensureDir(rootDir);

  const instanceDir = `${rootDir}/instances/${version}`;
  const accountDataFile = `${rootDir}/accounts.json`;

  const options: VersionOptions = {
    rootDir: rootDir,
    instanceDir: instanceDir,
    fabric: flags.fabric,
    jvmPath: flags.java ?? "java",
  };

  if (!flags.username) {
    if (await exists(accountDataFile)) {
      const refreshToken = JSON.parse(
        await Deno.readTextFile(accountDataFile),
      ).refresh;
      options.auth = await getAuthData(refreshToken);
    } else {
      options.auth = await getAuthData();
    }
    await Deno.writeTextFile(
      accountDataFile,
      JSON.stringify({
        refresh: options.auth.refresh,
      }),
    );
  } else {
    options.offlineUsername = flags.username;
  }

  const instance = await installVersion(version, options).catch((err) => {
    console.log(`An error occurred while installing: ${err.message}`);
    Deno.exit(1);
  });
  console.log(`Starting Minecraft ${version}...`);
  try {
    run(instance, options);
  } catch (err) {
    console.log(`An error occurred while running: ${err.message}`);
    Deno.exit(1);
  }
}

if (import.meta.main) {
  const VERSION = "0.6.0";
  await new Command()
    .name("cmd-launcher")
    .version(VERSION)
    .description("A minimal command line Minecraft launcher.")
    .command("launch", "Launch the game with the specified options.")
    .arguments("<version:string>")
    .option("-f, --fabric", "Run game with Fabric.")
    .option("-q, --quilt", "Run game with Quilt.")
    .option("-j, --java <path:string>", "Use the specified Java executable.")
    .option(
      "-u, --username <username:string>",
      "Use the specified offline mode username.",
    )
    .option(
      "-s, --server <server:string>",
      "Join the specified server on launch",
    )
    .action(async (options, ...args) => {
      await launchGame(options, ...args);
    })

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
