import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getStoreIdFromSession, getMetaTokens } from "@/lib/session";
import { SESSION_COOKIE_NAME } from "@/lib/constants";

const GRAPH_API_BASE = "https://graph.facebook.com/v25.0";

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE_NAME)?.value;
  if (!token || !getStoreIdFromSession(token)) {
    return NextResponse.json({ error: "Not authenticated" }, { status: 401 });
  }

  const metaTokens = getMetaTokens(token);
  if (!metaTokens?.facebookPageAccessToken) {
    return NextResponse.json({ error: "No FB token" }, { status: 400 });
  }

  const fbToken = metaTokens.facebookPageAccessToken;

  // Check permissions
  const permRes = await fetch(`${GRAPH_API_BASE}/me/permissions?access_token=${fbToken}`);
  const permData = permRes.ok ? await permRes.json() : null;

  // Check /me/accounts
  const acctRes = await fetch(`${GRAPH_API_BASE}/me/accounts?fields=id,name,access_token&access_token=${fbToken}`);
  const acctData = acctRes.ok ? await acctRes.json() : { error: await acctRes.text() };

  // Check /me
  const meRes = await fetch(`${GRAPH_API_BASE}/me?fields=id,name&access_token=${fbToken}`);
  const meData = meRes.ok ? await meRes.json() : null;

  // Try fetching the known page with instagram_business_account
  const knownPageId = process.env.META_PAGE_ID || "1032385586634860";
  const pageRes = await fetch(`${GRAPH_API_BASE}/${knownPageId}?fields=id,name,access_token,instagram_business_account{id,username,name}&access_token=${fbToken}`);
  const pageData = pageRes.ok ? await pageRes.json() : { error: await pageRes.text() };

  return NextResponse.json({
    storedTokens: {
      pageId: metaTokens.facebookPageId,
      pageName: metaTokens.facebookPageName,
      userId: metaTokens.facebookUserId,
      igUserId: metaTokens.instagramUserId,
      igUsername: metaTokens.instagramUsername,
      hasToken: !!metaTokens.facebookPageAccessToken,
    },
    permissions: permData,
    accounts: acctData,
    me: meData,
    pageWithIg: pageData,
  });
}
