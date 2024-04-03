export async function getFabricMeta(gameVersion: string) {
  const versionList = await (
    await fetch(`https://meta.fabricmc.net/v2/versions/loader/${gameVersion}`)
  ).json();

  return (
    await fetch(
      `https://meta.fabricmc.net/v2/versions/loader/${gameVersion}/${versionList[0].loader.version}/profile/json`,
    )
  ).json();
}

export async function getQuiltMeta(gameVersion: string) {
  return (
    await fetch(
      `https://meta.quiltmc.org/v3/versions/loader/${gameVersion}/0.24.0/profile/json`,
    )
  ).json();
}
