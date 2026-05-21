import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { createSession, setMetaTokens } from "@/lib/session";
import { SESSION_COOKIE_NAME, SESSION_TTL_MS } from "@/lib/constants";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";

const META_APP_ID = process.env.META_APP_ID || "";
const META_APP_SECRET = process.env.META_APP_SECRET || "";

export async function GET(request: NextRequest) {
  const cookieStore = await cookies();

  const code = request.nextUrl.searchParams.get("code");
  const state = request.nextUrl.searchParams.get("state");
  const error = request.nextUrl.searchParams.get("error");

  if (error) {
    const errorDesc =
      request.nextUrl.searchParams.get("error_description") || error;
    return NextResponse.redirect(
      new URL(
        `/seller/login?error=${encodeURIComponent(errorDesc)}`,
        request.url,
      ),
    );
  }

  if (!code) {
    return NextResponse.redirect(
      new URL("/seller/login?error=No+authorization+code", request.url),
    );
  }

  // Verify CSRF state
  const savedState = cookieStore.get("ig_oauth_state")?.value;
  if (!savedState || savedState !== state) {
    return NextResponse.redirect(
      new URL("/seller/login?error=Invalid+state", request.url),
    );
  }

  const redirectUri = `${request.nextUrl.origin}/api/auth/meta/instagram/callback`;

  try {
    // Step 1: Exchange code for short-lived access token
    const tokenRes = await fetch(
      "https://api.instagram.com/oauth/access_token",
      {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({
          client_id: META_APP_ID,
          client_secret: META_APP_SECRET,
          grant_type: "authorization_code",
          redirect_uri: redirectUri,
          code,
        }),
      },
    );

    if (!tokenRes.ok) {
      const err = await tokenRes.json();
      throw new Error(
        err.error_message || err.error?.message || "Token exchange failed",
      );
    }
    const tokenData = await tokenRes.json();
    const shortLivedToken = tokenData.access_token;
    const userId = tokenData.user_id;

    // Step 2: Exchange for long-lived token
    const longLivedRes = await fetch(
      `https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=${META_APP_SECRET}&access_token=${shortLivedToken}`,
    );

    let accessToken = shortLivedToken;
    if (longLivedRes.ok) {
      const longLivedData = await longLivedRes.json();
      accessToken = longLivedData.access_token;
    }

    // Step 3: Get user profile
    const profileRes = await fetch(
      `https://graph.instagram.com/v25.0/me?fields=user_id,username,name&access_token=${accessToken}`,
    );
    let username = "";
    let displayName = "";
    if (profileRes.ok) {
      const profileData = await profileRes.json();
      username = profileData.username || "";
      displayName = profileData.name || username;
    }

    // Step 4: Find or create store in Convex
    const store = await convexServer.mutation(
      api.stores.findOrCreateByInstagram,
      {
        instagramUserId: String(userId),
        instagramUsername: username,
        name: displayName || username || "My Store",
      },
    );

    // Step 5: Create session
    const sessionToken = createSession(store.storeId as string);

    // Step 6: Store Meta tokens
    setMetaTokens(sessionToken, {
      instagramAccessToken: accessToken,
      instagramUserId: String(userId),
      instagramUsername: username,
    });

    // Step 7: Redirect to seller dashboard
    const response = NextResponse.redirect(
      new URL("/seller", request.url),
    );

    response.cookies.set(SESSION_COOKIE_NAME, sessionToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    response.cookies.set("seller_logged_in", "1", {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    response.cookies.delete("ig_oauth_state");

    // Store slug in a client-readable cookie for localStorage sync
    response.cookies.set("store_slug", store.slug, {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    response.cookies.set("store_name", store.name, {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    return response;
  } catch (err) {
    console.error("Instagram OAuth callback error:", err);
    const message = err instanceof Error ? err.message : "OAuth failed";
    return NextResponse.redirect(
      new URL(
        `/seller/login?error=${encodeURIComponent(message)}`,
        request.url,
      ),
    );
  }
}
