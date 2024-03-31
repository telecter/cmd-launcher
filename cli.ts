import { exists } from "https://deno.land/std@0.219.1/fs/mod.ts";
import { parseArgs } from "https://deno.land/std@0.220.1/cli/mod.ts";
import * as api from "./api/game.ts";
import { getAuthData } from "./api/auth.ts";
import {
  fetchFabricLibraries,
  getFabricMeta,
  getQuiltMeta,
} from "./api/fabric.ts";

function printHelp() {
  console.log(`
usage: cmd-launcher [...options]
A minimal command line Minecraft launcher.

Options:
  -l, --launch      Launch a specific version of the game
  -u, --usernam     Set the offline mode username
  --fabric      Launch the game with the Fabric mod loader
  --quilt       Launch the game with the Quilt mod loader
  -s, --server      Join the specified server on launch
  -h, --help        Show this help and exit`);
}

async function main(args: string[]) {
  api.registerDownloadListener((url) => {
    console.log(`Downloading ${url}`);
  });
  const flags = parseArgs(args, {
    string: ["launch", "server", "username"],
    boolean: ["help", "fabric", "quilt"],
    alias: { help: "help", launch: "l", server: "s", username: "u" },
    unknown: (arg) => {
      console.log(`Unknown argument: ${arg}`);
      printHelp();
      Deno.exit(1);
    },
  });
  let version = flags.launch ?? null;

  if (flags.help) {
    printHelp();
    Deno.exit();
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

  let libraryPaths = await api
    .fetchLibraries(data.libraries, rootDir)
    .catch((err) => {
      console.error(
        `An error occurred while downloading a library: ${err.message}`,
      );
      Deno.exit(1);
    });

  if (flags.fabric || flags.quilt) {
    const fabricData = flags.quilt
      ? await getQuiltMeta(version)
      : await getFabricMeta(version);
    const paths = await fetchFabricLibraries(fabricData.libraries, rootDir);
    libraryPaths = [...libraryPaths, ...paths];
    mainClass = fabricData.mainClass;
  }

  const assets = await api.getAndSaveAssetData(data, rootDir);
  await api.fetchAssets(assets, rootDir).catch((err) => {
    console.error(
      `An error occurred while downloading an asset: ${err.message}`,
    );
    Deno.exit(1);
  });

  const clientPath = await api.fetchClient(data, instanceDir).catch((err) => {
    console.error(`Failed to download Minecraft client: ${err.message}`);
    Deno.exit(1);
  });

  const classPath = [clientPath, ...libraryPaths];
  const javaArgs = ["-cp", classPath.join(":")];

  const [accessToken, username, uuid] = flags.username
    ? ["abc", flags.username, crypto.randomUUID()]
    : await getAuthData(`${rootDir}/accounts.json`);

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
