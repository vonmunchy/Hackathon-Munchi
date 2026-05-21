import { NextRequest, NextResponse } from "next/server";

const META_APP_ID = process.env.META_APP_ID || "";

export async function GET(request: NextRequest) {
  const redirectUri = `${request.nextUrl.origin}/api/auth/meta/instagram/callback`;

  // Instagram API with Instagram Login scopes
  const scopes = [
    "instagram_business_basic",
    "instagram_business_manage_messages",
    "instagram_business_manage_comments",
    "instagram_business_content_publish",
  ].join(",");

  const state = crypto.randomUUID();

  const authUrl = new URL("https://www.instagram.com/oauth/authorize");
  authUrl.searchParams.set("client_id", META_APP_ID);
  authUrl.searchParams.set("redirect_uri", redirectUri);
  authUrl.searchParams.set("scope", scopes);
  authUrl.searchParams.set("state", state);
  authUrl.searchParams.set("response_type", "code");
  authUrl.searchParams.set("enable_fb_login", "0");
  authUrl.searchParams.set("force_authentication", "1");

  const response = NextResponse.redirect(authUrl.toString());
  response.cookies.set("ig_oauth_state", state, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 600,
  });

  return response;
}
