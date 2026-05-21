import { NextRequest, NextResponse } from "next/server";

const META_APP_ID = process.env.META_APP_ID || "";
const META_FB_CONFIG_ID = process.env.META_FB_CONFIG_ID || "992927346546138";
const GRAPH_API_VERSION = "v25.0";

export async function GET(request: NextRequest) {
  const redirectUri = `${request.nextUrl.origin}/api/auth/meta/facebook/callback`;

  const state = crypto.randomUUID();

  const authUrl = new URL(
    `https://www.facebook.com/${GRAPH_API_VERSION}/dialog/oauth`,
  );
  authUrl.searchParams.set("client_id", META_APP_ID);
  authUrl.searchParams.set("redirect_uri", redirectUri);
  authUrl.searchParams.set("config_id", META_FB_CONFIG_ID);
  authUrl.searchParams.set("state", state);
  authUrl.searchParams.set("response_type", "code");
  // Force re-authorization to pick up new page permissions
  authUrl.searchParams.set("auth_type", "rerequest");

  const response = NextResponse.redirect(authUrl.toString());
  response.cookies.set("meta_oauth_state", state, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 600,
  });

  return response;
}
