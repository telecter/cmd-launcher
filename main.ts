import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.220.1/cli/mod.ts";
import * as api from "./api/game.ts";
import { getAuthData } from "./api/auth.ts";
import { fetchFabricLibraries, getFabricMeta } from "./api/fabric.ts";
import { fetchClient } from "./api/game.ts";

async function update() {
  if (Deno.execPath().includes("deno")) {
    throw Error("Cannot update non-executable");
  }
  const tags = await (
    await fetch("https://api.github.com/repos/telectr/cmd-launcher/tags")
  ).json();
  const latestTag = tags[0].name;
  console.log(`Upgrading to ${latestTag}`);
  const data = await (
    await fetch(
      `https://github.com/telectr/cmd-launcher/releases/download/${latestTag}/launcher-${Deno.build.target}`,
    )
  ).arrayBuffer();
  await Deno.writeFile(Deno.execPath(), new Uint8Array(data));
  Deno.exit();
}

function printHelp() {
  console.log(`
usage: cmd-launcher [...options]
A minimal command line Minecraft launcher.

Options:
  -l, --launch      Launch a specific version of the game
  -f, --fabric      Launch the game with the Fabric mod loader
  -s, --server      Join the specified server on launch

  --update          Update the launcher
  -h, --help        Show this help and exit`);
}

async function main(args: string[]) {
  const flags = parseArgs(args, {
    string: ["version", "launch", "server"],
    boolean: ["help", "update", "fabric"],
    alias: { help: "help", launch: "l", server: "s" },
    unknown: (arg) => {
      console.log(`Unknown argument: ${arg}`, "ERROR");
      printHelp();
      Deno.exit(1);
    },
  });
  let version = flags.launch ?? null;

  if (flags.help) {
    printHelp();
    Deno.exit();
  }
  if (flags.update) {
    await update();
  }

  const data = await api.getVersionData(version).catch((err) => {
    console.log(`Failed to get version data: ${err.message}`, "ERROR");
    Deno.exit(1);
  });

  version = data.id;
  let mainClass = data.mainClass;

  const rootDir = `${Deno.cwd()}/minecraft`;
  const instanceDir = `${rootDir}/instances/${flags.name ?? version}`;

  if (!(await exists(rootDir))) {
    await Deno.mkdir(rootDir);
  }
  if (!(await exists(instanceDir))) {
    await Deno.mkdir(instanceDir, { recursive: true });
  }

  Deno.chdir(rootDir);

  console.log("Downloading libraries...");
  let libraryPaths = await api.fetchLibraries(data.libraries, rootDir);

  if (flags.fabric) {
    const fabricData = await getFabricMeta(version);
    const paths = await fetchFabricLibraries(fabricData.libraries, rootDir);
    libraryPaths = [...libraryPaths, ...paths];
    mainClass = fabricData.mainClass;
  }

  console.log("Downloading assets...");

  const assets = await api.getAndSaveAssetData(data, rootDir);
  await api.fetchAssets(assets, rootDir);

  const clientPath = await fetchClient(data, rootDir);

  const classPath = [clientPath, ...libraryPaths];

  const javaArgs = ["-cp", classPath.join(":")];

  const [accessToken, username, uuid] = await getAuthData(
    `${rootDir}/accounts.json`,
  );

  const gameArgs = [
    "--version",
    "",
    "--accessToken",
    accessToken,
    "--assetsDir",
    `${rootDir}/assets`,
    "--assetIndex",
    data.assetIndex.id,
    "--gameDir",
    instanceDir,
    "--uuid",
    uuid,
    "--username",
    username,
  ];
  if (Deno.build.os == "darwin") {
    javaArgs.push("-XstartOnFirstThread");
  }
  if (flags.fabric) {
    javaArgs.push("-DFabricMcEmu= net.minecraft.client.main.Main ");
  }
  if (flags.server) {
    gameArgs.push("--quickPlayMultiplayer", flags.server);
  }

  console.log(`Starting Minecraft ${version}...`);

  Deno.chdir(instanceDir);
  new Deno.Command("java", {
    args: [...javaArgs, mainClass, ...gameArgs],
  }).spawn();
}

if (import.meta.main) {
  main(Deno.args);
}
