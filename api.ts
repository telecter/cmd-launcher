import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";
import {  Asset, Library, VersionManifest } from "./types.ts";
import { download } from "./util.ts";
import { AssetData, VersionData } from "./types.ts";

export async function getVersionManifest() {
  return <VersionManifest>(await (await fetch("https://launchermeta.mojang.com/mc/game/version_manifest.json")).json())
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
    return <VersionData>await (await fetch(release.url)).json()
}

export async function getAssetData(url: string) {
  return <AssetData>await (await fetch(url)).json()
}

export async function downloadLibrary(library: Library, rootDir: string) {
    const artifact = library.downloads.artifact
    const path = `${rootDir}/${artifact.path}`
    if (!await exists(path)) {
      await download(artifact.url, path)
    }
}

export async function downloadAsset(asset: Asset, rootDir: string) {
      const objectPath = `${asset.hash.slice(0, 2)}/${asset.hash}`
      const path = `${rootDir}/objects/${objectPath}`
      if (!await exists(path)) {
        await download(`https://resources.download.minecraft.net/${objectPath}`, path)
      }
}

export async function getAuthToken() {
  const client_id = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
  const scope = "XboxLive.SignIn"
  const redirect_uri = "http://localhost:8000/signin"
  const url = new URL("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
  url.search = new URLSearchParams({
    "client_id": client_id,
    "response_type": "code",
    "redirect_uri": redirect_uri,
    "scope": scope,
    "response_mode": "query"
  }).toString()

  new Deno.Command("open", { args: [url.toString()] }).spawn()

  let authCode: string

  const server = Deno.serve((req) => {
    const url = new URL(req.url)
    if (url.pathname == "/signin") {
      authCode = <string>url.searchParams.get("code")
      queueMicrotask(server.shutdown)
      return new Response("Response recorded", { status: 200 })
    }
    return new Response("Not Found", { status: 404 })
  })
  await server.finished

  const authTokenData = await (await fetch("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", {
    method: "POST",
    body: new URLSearchParams({
      "client_id": client_id,
      "scope": scope,
      "redirect_uri": redirect_uri,
      "grant_type": "authorization_code",
      "code": authCode!
    }),
    headers: { "Content-Type": "application/x-www-form-urlencoded" }
  })).json()

  const authToken = authTokenData.access_token

  const xboxAuthData = await (await fetch("https://user.auth.xboxlive.com/user/authenticate", {
    method: "POST",
    body: JSON.stringify({
        Properties: {
        AuthMethod: "RPS",
        SiteName: "user.auth.xboxlive.com",
        RpsTicket: `d=${authToken}`
      },
      RelyingParty: "http://auth.xboxlive.com",
      TokenType: "JWT"
    }),
    headers: {
      "Content-Type": "application/json",
      "Accept": "application/json"
    }
  })).json()
  const xblToken = xboxAuthData.Token
  const userhash = xboxAuthData.DisplayClaims.xui[0].uhs

  const xstsData = await (await fetch("https://xsts.auth.xboxlive.com/xsts/authorize", {
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
  const xstsToken = xstsData.Token
  const loginXboxData = await fetch("https://api.minecraftservices.com/authentication/login_with_xbox", {
    method: "POST",
    body: JSON.stringify({
      identityToken: `XBL3.0 x=${userhash};${xstsToken}`
    }),
    headers: {
      "Content-Type": "application/json",
      "Accept": "application/json"
    }
  })
  console.log(await loginXboxData.json())
  console.log(loginXboxData.status, loginXboxData.statusText)
}