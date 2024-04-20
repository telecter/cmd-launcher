import { isAbsolute } from "https://deno.land/std@0.221.0/path/is_absolute.ts";
import {
  Command,
  GithubProvider,
  UpgradeCommand,
} from "https://deno.land/x/cliffy@v1.0.0-rc.3/command/mod.ts";
import { launchCommand } from "./launch.ts";
import { installModsCommand } from "./install_mod.ts";

const VERSION = "0.6.1";

export function getDirs(version: string, dir?: string) {
  const rawRootDir = dir ?? `${Deno.env.get("HOME")}/.minecraft`;
  const rootDir = isAbsolute(rawRootDir)
    ? rawRootDir
    : `${Deno.cwd()}/${rawRootDir}`;
  return [rootDir, `${rootDir}/instances/${version}`];
}

if (import.meta.main) {
  const cmd = new Command()
    .name("cmd-launcher")
    .version(VERSION)
    .description("A minimal command line Minecraft launcher.")
    .globalOption(
      "-d, --dir <path:string>",
      "Set the root directory for the game.",
    )
    .command("launch", launchCommand)
    .command("mod", installModsCommand)
    .command(
      "upgrade",
      new UpgradeCommand({
        provider: new GithubProvider({ repository: "telectr/cmd-launcher" }),
        main: "cli.ts",
        args: ["-A"],
      }),
    );
  try {
    await cmd.parse();
  } catch (err) {
    console.error(err);
    console.error(
      `%cerror%c: ${err.message}`,
      "color: red; font-weight: bold;",
      "",
    );
  }
}
