import { exists } from "https://deno.land/std@0.219.1/fs/exists.ts";

const client_id = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7";
const scope = "XboxLive.SignIn";
const redirect_uri = "http://localhost:8000/signin";

async function getMsaAuthCode() {
  const url = new URL(
    "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
  );
  url.search = new URLSearchParams({
    client_id: client_id,
    response_type: "code",
    redirect_uri: redirect_uri,
    scope: scope,
    response_mode: "query",
  }).toString();

  new Deno.Command("open", { args: [url.toString()] }).spawn();

  let authCode: string;

  const server = Deno.serve((req) => {
    const url = new URL(req.url);
    if (url.pathname == "/signin") {
      authCode = <string> url.searchParams.get("code");
      queueMicrotask(server.shutdown);
      return new Response("Response recorded", { status: 200 });
    }
    return new Response("Not Found", { status: 404 });
  });
  await server.finished;
  return authCode!;
}

async function getMsaAuthToken(authCode: string): Promise<string> {
  const authTokenData = await (
    await fetch(
      "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
      {
        method: "POST",
        body: new URLSearchParams({
          client_id: client_id,
          scope: scope,
          redirect_uri: redirect_uri,
          grant_type: "authorization_code",
          code: authCode!,
        }),
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
      },
    )
  ).json();

  return authTokenData.access_token;
}

async function getXboxAuthData(msaAuthToken: string) {
  const xboxAuthData = await (
    await fetch("https://user.auth.xboxlive.com/user/authenticate", {
      method: "POST",
      body: JSON.stringify({
        Properties: {
          AuthMethod: "RPS",
          SiteName: "user.auth.xboxlive.com",
          RpsTicket: `d=${msaAuthToken}`,
        },
        RelyingParty: "http://auth.xboxlive.com",
        TokenType: "JWT",
      }),
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
    })
  ).json();
  console.log(xboxAuthData);
  return [xboxAuthData.Token, xboxAuthData.DisplayClaims.xui[0].uhs];
}

async function getXstsToken(xblToken: string) {
  const xstsData = await (
    await fetch("https://xsts.auth.xboxlive.com/xsts/authorize", {
      method: "POST",
      body: JSON.stringify({
        Properties: {
          SandboxId: "RETAIL",
          UserTokens: [xblToken],
        },
        RelyingParty: "rp://api.minecraftservices.com/",
        TokenType: "JWT",
      }),
    })
  ).json();
  console.log(xstsData);
  return xstsData.Token;
}

async function getMinecraftAuthToken(xstsToken: string, userhash: string) {
  const loginXboxData = await (
    await fetch(
      "https://api.minecraftservices.com/authentication/login_with_xbox",
      {
        method: "POST",
        body: JSON.stringify({
          identityToken: `XBL3.0 x=${userhash};${xstsToken}`,
        }),
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
      },
    )
  ).json();
  return loginXboxData.access_token;
}

async function getProfileData(jwtToken: string) {
  const accountOwnershipData = await fetch(
    "https://api.minecraftservices.com/entitlements/mcstore",
    {
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    },
  );
  const profileData = await (
    await fetch("https://api.minecraftservices.com/minecraft/profile", {
      headers: {
        Authorization: `Bearer ${jwtToken}`,
      },
    })
  ).json();
  return [profileData.name, profileData.id];
}

export async function getAuthData(rootDir: string) {
  if (await exists(`${rootDir}/accounts.json`)) {
  }

  const authCode = await getMsaAuthCode();
  const msaAuthToken = await getMsaAuthToken(authCode);
  const [xblToken, userhash] = await getXboxAuthData(msaAuthToken);
  const xstsToken = await getXstsToken(xblToken);
  const jwtToken = await getMinecraftAuthToken(xstsToken, userhash);
  const [username, uuid] = await getProfileData(jwtToken);
  await Deno.writeTextFile(
    `${rootDir}/accounts.json`,
    JSON.stringify({
      username: username,
    }),
  );
  return [jwtToken, username, uuid];
}
