import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.220.1/cli/mod.ts";
import * as api from "./api/mojang.ts";
import { log, download, writeOnLine } from "./util.ts";
import { getAuthData } from "./api/auth.ts";
import { fetchFabricLibrary, getFabricMeta } from "./mod_loaders/fabric.ts";

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
  A command line Minecraft launcher.

  Options:
    -l, --launch      Launch a specific version of the game
    -f, --fabric      Launch the game with the Fabric mod loader
    -s, --server      Join the specified server on launch

    --update          Update the launcher
    -h, --help        Show this help and exit
  `);
}

async function main(args: string[]) {
  const flags = parseArgs(args, {
    string: ["version", "launch", "server"],
    boolean: ["help", "update", "fabric"],
    alias: { help: "help", launch: "l", server: "s" },
    unknown: (arg) => {
      console.log(`[Invalid argument: ${arg}`);
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
    log(`Failed to get version data: ${err.message}`, "ERROR");
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

  const libraryPathList = [];

  log("Downloading libraries...");
  for (const library of data.libraries) {
    const path = await api.fetchLibrary(library, rootDir);
    libraryPathList.push(path);
    writeOnLine(library.downloads.artifact.path);
  }

  if (flags.fabric) {
    log("Loading Fabric...");
    const fabricData = await getFabricMeta(version);
    for (const library of fabricData.libraries) {
      const libraryPath = await fetchFabricLibrary(library, rootDir);
      libraryPathList.push(libraryPath);
    }
    mainClass = fabricData.mainClass;
  }

  log("Downloading assets...");
  const assets = await api.getAndSaveAssetData(data.assetIndex, rootDir);
  let i = 0;
  const numOfAssets = Object.keys(assets).length;
  for (const asset of Object.values(assets)) {
    i++;
    await api.fetchAsset(asset, rootDir);
    writeOnLine(`${i}/${numOfAssets}`);
  }

  const clientPath = `${instanceDir}/client.jar`;
  if (!(await exists(clientPath))) {
    log("Downloading client...");
    await download(data.downloads.client.url, clientPath);
  }

  const classPath = [clientPath, ...libraryPathList];

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

  log(`Starting Minecraft ${version}...`);

  Deno.chdir(instanceDir);
  new Deno.Command("java", {
    args: [...javaArgs, mainClass, ...gameArgs],
  }).spawn();
}

if (import.meta.main) {
  main(Deno.args);
}
