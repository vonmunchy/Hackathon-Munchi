import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getMetaTokens, getMetaTokensFromCookie } from "@/lib/session";
import { SESSION_COOKIE_NAME } from "@/lib/constants";

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE_NAME)?.value;

  if (!token) {
    // Check if we have meta_tokens cookie (session may have been lost but cookies persist)
    const metaTokensCookie = cookieStore.get("meta_tokens")?.value;
    if (!metaTokensCookie) {
      return NextResponse.json(
        { error: "Authentication required" },
        { status: 401 },
      );
    }
  }

  // Try in-memory first, then fall back to cookie
  let metaTokens = token ? getMetaTokens(token) : null;
  if (!metaTokens) {
    metaTokens = getMetaTokensFromCookie(
      cookieStore.get("meta_tokens")?.value,
    );
  }

  return NextResponse.json({
    facebook: metaTokens?.facebookPageAccessToken
      ? {
          connected: true,
          pageId: metaTokens.facebookPageId,
          pageName: metaTokens.facebookPageName,
        }
      : { connected: false },
    instagram: metaTokens?.instagramAccessToken
      ? {
          connected: true,
          userId: metaTokens.instagramUserId,
          username: metaTokens.instagramUsername,
        }
      : { connected: false },
  });
}
