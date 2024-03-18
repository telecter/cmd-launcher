import { type Asset, type Library, type VersionManifest } from "./types.ts";
import { download } from "./util.ts";

export async function getVersionManifest() {
  return <VersionManifest>(await (await fetch("https://launchermeta.mojang.com/mc/game/version_manifest.json")).json())
}

export async function filterVersions(filter: "release"|"snapshot") {
  return (await getVersionManifest()).versions.filter((element) => element.type == filter)
}


export async function getVersionData(version: string|null) {
    const data = await getVersionManifest()
    if (!version) {
      version = data.latest.release
    }
    const release = data.versions.find((element) => element.id == version)
    if (!release) {
      throw Error("Invalid version")
    }
    return (await fetch(release.url)).json()
}

export async function downloadAssetData(url: string, id: string, rootDir: string) {
  await download(url, `${rootDir}/assets/indexes/${id}.json`)
}

export async function downloadLibrary(library: Library, rootDir: string) {
    const artifact = library.downloads.artifact
    await download(artifact.url, `${rootDir}/libraries/${artifact.path}`)
}
export async function downloadAsset(asset: Asset, rootDir: string) {
      const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`
      const path = `${rootDir}/assets/objects/${objectPath}`
      await download(`https://resources.download.minecraft.net/${objectPath}`, path)
}

export async function getAuthToken() {
  const client_id = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
  const url = new URL("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
  const params = new URLSearchParams({
    "client_id": client_id,
    "response_type": "code",
    "redirect_uri": "http://localhost:8000/signin",
    "scope": "XboxLive.SignIn",
    "response_mode": "query"
  })
  url.search = params.toString()

  new Deno.Command("open", {
    args: [url.toString()]
  }).spawn()

  let authCode: string

  const server = Deno.serve((req) => {
    const url = new URL(req.url)
    if (url.pathname == "/signin") {
      authCode = <string>url.searchParams.get("code")
      console.log(authCode)
      queueMicrotask(server.shutdown)
      return new Response("Response recorded", { status: 200 })
    }
    return new Response("Not Found", { status: 404 })
  })
  await server.finished

  const tokenJson = await (await fetch("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", {
    method: "POST",
    body: new URLSearchParams({
      "client_id": "6a533aa3-afbf-45a4-91bc-8c35a37e35c7",
      "scope": "XboxLive.SignIn",
      "redirect_uri": "http://localhost:8000/signin",
      "grant_type": "authorization_code",
      "code": authCode!
    }),
    headers: { "Content-Type": "application/x-www-form-urlencoded" }
  })).json()
  const token = tokenJson.access_token

  const xboxJson = await (await fetch("https://user.auth.xboxlive.com/user/authenticate", {
    method: "POST",
    body: JSON.stringify({
      Properties: {
        AuthMethod: "RPS",
        SiteName: "user.auth.xboxlive.com",
        RpsTicket: `d=${token}`
      },
      RelyingParty: "http://auth.xboxlive.com",
      TokenType: "JWT"
    }),
    headers: {
      "Content-Type": "application/json",
      "Accept": "application/json"
    }
  })).json()
  const xblToken = xboxJson.Token
  const userhash = xboxJson.DisplayClaims.xui[0].uhs
  console.log(`USERHASH: ${userhash}`)
  const xstsJson = await (await fetch("https://xsts.auth.xboxlive.com/xsts/authorize", {
    method: "POST",
    body: JSON.stringify({
      Properties: {
        SandboxId: "RETAIL",
        UserTokens: [
          xblToken
        ]
      },
      RelyingParty: "rp://api.minecraftservices.com/",
      TokenType: "JWT"
    })
  })).json()
  const xstsToken = xstsJson.Token
  const data = JSON.stringify({
    identityToken: `XBL3.0 x=${userhash};${xstsToken}`
  })
  console.log(data)
  const loginXboxJson = await fetch("https://api.minecraftservices.com/authentication/login_with_xbox", {
    method: "POST",
    body: data
  })
  console.log(loginXboxJson.status, loginXboxJson.statusText)
}
await getAuthToken()