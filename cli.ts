import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { Command } from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/mod.ts";
import { Version } from "./launcher.ts";
import { getAuthData } from "./api/auth.ts";

type LaunchOptions = {
  fabric?: boolean;
  quilt?: boolean;
  java?: string;
  server?: string;
  username?: string;
};

async function launchGame(flags: LaunchOptions, ...args: string[]) {
  const version = args[0];

  const rootDir = `${Deno.cwd()}/minecraft`;
  const instanceDir = `${rootDir}/instances/${version}`;

  if (!(await exists(rootDir))) {
    await Deno.mkdir(rootDir);
  }
  if (!(await exists(instanceDir))) {
    await Deno.mkdir(instanceDir, { recursive: true });
  }

  const [accessToken, username, uuid] = flags.username
    ? ["a", flags.username, crypto.randomUUID()]
    : await getAuthData(`${rootDir}/accounts.json`);
  const a = new Version(version, {
    rootDir: rootDir,
    accessToken: accessToken,
    uuid: uuid,
    username: username,
    instanceDir: instanceDir,
    fabric: flags.fabric,
  });
  await a.init();

  await a.install((url) => console.log(`Downloading ${url}`));
  a.run();
}

if (import.meta.main) {
  const VERSION = "0.6.0";
  await new Command()
    .name("cmd-launcher")
    .version(VERSION)
    .description("A minimal command line Minecraft launcher.")
    .command("launch", "Launch the game with the specified options.")
    .arguments("<version:string>")
    .option("--fabric, -f [fabric:boolean]", "Use the Fabric mod loader")
    .option("--quilt, -q [quilt:boolean]", "Use the Quilt mod loader")
    .option("--java, -j <java:string>", "Use the specified Java executable.")
    .option(
      "--username, -u <username:string>",
      "Use the specified offline mode username.",
    )
    .option(
      "--server, -s <server:string>",
      "Join the specified server on launch",
    )
    .action(async (options, ...args) => {
      await launchGame(options, ...args);
    })
    .parse(Deno.args);
}
