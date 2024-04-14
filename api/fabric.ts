import { fetchJSONData } from "../util.ts";

async function getLoaderMeta(urlPrefix: string, gameVersion: string) {
  const loaderVersions = await fetchJSONData(`${urlPrefix}/${gameVersion}`);
  const meta = await fetchJSONData(
    `${urlPrefix}/${gameVersion}/${loaderVersions[0].loader.version}/profile/json`,
  );
  return meta;
}

export async function getFabricMeta(gameVersion: string) {
  const meta = await getLoaderMeta(
    "https://meta.fabricmc.net/v2/versions/loader",
    gameVersion,
  );
  return meta;
}

export async function getQuiltMeta(gameVersion: string) {
  const meta = await getLoaderMeta(
    "https://meta.quiltmc.org/v3/versions/loader",
    gameVersion,
  );
  return meta;
}
