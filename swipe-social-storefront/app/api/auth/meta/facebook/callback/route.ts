import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { createSession, setMetaTokens, encodeMetaTokensForCookie, MetaTokens } from "@/lib/session";
import { SESSION_COOKIE_NAME, SESSION_TTL_MS } from "@/lib/constants";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";

const META_APP_ID = process.env.META_APP_ID || "";
const META_APP_SECRET = process.env.META_APP_SECRET || "";
const GRAPH_API_VERSION = "v25.0";
const GRAPH_API_BASE = `https://graph.facebook.com/${GRAPH_API_VERSION}`;

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
  const savedState = cookieStore.get("meta_oauth_state")?.value;
  if (!savedState || savedState !== state) {
    return NextResponse.redirect(
      new URL("/seller/login?error=Invalid+state", request.url),
    );
  }

  const redirectUri = `${request.nextUrl.origin}/api/auth/meta/facebook/callback`;

  try {
    // Step 1: Exchange code for user access token
    const tokenUrl = new URL(`${GRAPH_API_BASE}/oauth/access_token`);
    tokenUrl.searchParams.set("client_id", META_APP_ID);
    tokenUrl.searchParams.set("client_secret", META_APP_SECRET);
    tokenUrl.searchParams.set("redirect_uri", redirectUri);
    tokenUrl.searchParams.set("code", code);

    const tokenRes = await fetch(tokenUrl.toString());
    if (!tokenRes.ok) {
      const err = await tokenRes.json();
      throw new Error(err.error?.message || "Token exchange failed");
    }
    const tokenData = await tokenRes.json();
    const userAccessToken = tokenData.access_token;

    // Step 2: Get user info
    const meRes = await fetch(
      `${GRAPH_API_BASE}/me?fields=id,name,email&access_token=${userAccessToken}`,
    );
    if (!meRes.ok) {
      throw new Error("Failed to fetch user info");
    }
    const meData = await meRes.json();

    // Step 3: Get the user's Facebook Pages
    const pagesRes = await fetch(
      `${GRAPH_API_BASE}/me/accounts?fields=id,name,access_token&access_token=${userAccessToken}`,
    );
    const pagesData = pagesRes.ok ? await pagesRes.json() : { data: [] };
    let pages = pagesData.data || [];

    // Fallback: if /me/accounts returns empty (common in dev mode with Facebook Login
    // for Business), try fetching the known page directly using its ID
    if (pages.length === 0) {
      const envPageId = process.env.META_PAGE_ID;
      if (envPageId) {
        const directRes = await fetch(
          `${GRAPH_API_BASE}/${envPageId}?fields=id,name,access_token&access_token=${userAccessToken}`,
        );
        if (directRes.ok) {
          const directPage = await directRes.json();
          if (directPage.id) {
            pages = [directPage];
          }
        }
      }
    }

    const page = pages[0];

    // Step 4: Find or create store in Convex
    const store = await convexServer.mutation(
      api.stores.findOrCreateByFacebook,
      {
        facebookUserId: meData.id,
        facebookPageId: page?.id,
        name: page?.name || meData.name || "My Store",
      },
    );

    // Step 5: Fetch linked Instagram Business Account from the Page
    let igAccountId = "";
    let igUsername = "";
    if (page?.id && page?.access_token) {
      const igRes = await fetch(
        `${GRAPH_API_BASE}/${page.id}?fields=instagram_business_account{id,username}&access_token=${page.access_token}`,
      );
      if (igRes.ok) {
        const igData = await igRes.json();
        if (igData.instagram_business_account) {
          igAccountId = igData.instagram_business_account.id;
          igUsername = igData.instagram_business_account.username || "";
        }
      }
    }

    // Step 6: Create session (in-memory for Meta tokens + Convex for multi-tenant auth)
    const sessionToken = createSession(store.storeId as string);

    // Create persistent session in Convex for multi-tenant isolation
    await convexServer.mutation(api.sessions.create, {
      token: sessionToken,
      storeId: store.storeId as any,
    });

    // Step 7: Store Meta tokens in session AND cookie (cookie survives server restarts)
    const pageToken = page?.access_token || userAccessToken;
    const metaTokensData: MetaTokens = {
      facebookPageAccessToken: pageToken,
      facebookPageId: page?.id || meData.id,
      facebookPageName: page?.name || meData.name || "My Page",
      facebookUserId: meData.id,
      ...(igAccountId && {
        instagramAccessToken: pageToken,
        instagramUserId: igAccountId,
        instagramUsername: igUsername,
      }),
    };
    setMetaTokens(sessionToken, metaTokensData);

    // Step 8: Redirect to seller dashboard
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

    // Client-readable session token for Convex multi-tenant auth
    response.cookies.set("seller_session_token", sessionToken, {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    response.cookies.delete("meta_oauth_state");

    // Client-readable cookies for localStorage sync
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

    // Check onboarding status and set cookie
    const fullStore = await convexServer.query(api.stores.getBySlug, { slug: store.slug });
    if (fullStore?.onboardingComplete) {
      response.cookies.set("onboarding_complete", "1", {
        httpOnly: false,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        path: "/",
        maxAge: SESSION_TTL_MS / 1000,
      });
    }

    // Store meta tokens in httpOnly cookie (survives server restarts in dev mode)
    response.cookies.set("meta_tokens", encodeMetaTokensForCookie(metaTokensData), {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    return response;
  } catch (err) {
    console.error("Facebook OAuth callback error:", err);
    const message = err instanceof Error ? err.message : "OAuth failed";
    return NextResponse.redirect(
      new URL(
        `/seller/login?error=${encodeURIComponent(message)}`,
        request.url,
      ),
    );
  }
}
