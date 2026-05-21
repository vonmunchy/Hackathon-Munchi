---
title: "Meta Business OAuth Login and Social Publishing with Next.js + Convex"
date: 2026-05-21
category: integration-issues
module: meta-api-publishing
problem_type: integration_issue
component: authentication
symptoms:
  - "Facebook /me/accounts returns empty array despite correct OAuth permissions"
  - "In-memory session tokens lost on Next.js dev server hot reload"
  - "OAuth callback creates orphan store instead of linking to existing business store"
  - "Instagram shows not linked even when connected in Meta Business Suite"
  - "Meta APIs reject localhost image URLs for photo/Instagram publishing"
root_cause: incomplete_setup
resolution_type: code_fix
severity: high
tags:
  - meta-graph-api
  - facebook-login-for-business
  - oauth
  - instagram-publishing
  - nextjs
  - convex
  - session-persistence
  - convex-storage
---

# Meta Business OAuth Login and Social Publishing with Next.js + Convex

## Problem

Implementing a single "Continue with Meta Business" login for sellers that authenticates via Facebook Login for Business, connects both their Facebook Page and Instagram Business account, and enables direct product publishing to both platforms — with product images hosted on Convex Storage for public accessibility.

## Symptoms

- Facebook `/me/accounts` Graph API endpoint returned empty array after OAuth, even with `pages_show_list` permission granted
- Meta tokens stored in Next.js in-memory `Map` were lost on every dev server hot reload, breaking publishing until re-login
- OAuth callback created a new store with the seller's personal name instead of linking to the existing business store
- Instagram showed "not linked" on the Social page despite being linked in Meta Business Suite
- Facebook photo posts and Instagram publishing failed because product images were local paths (`/demo-products/abaya.png`) inaccessible to Meta's servers

## What Didn't Work

- **Passing `scope` parameter to Facebook Login for Business OAuth**: Facebook Login for Business ignores individual `scope` params and requires a `config_id` instead. Without `config_id`, only basic user permissions are granted — no page or Instagram access.
- **Relying on `/me/accounts` alone**: In dev mode with Standard Access, this endpoint returns empty even for app admins who have granted page permissions. The configuration must be created in the Meta Developer Console and referenced via `config_id`.
- **In-memory session Map for Meta tokens**: Next.js dev mode rebuilds modules on file changes, resetting all module-level state. Any file edit wiped the session Map, requiring a full re-login through OAuth.
- **Matching OAuth users by `facebookUserId` first**: The user's personal Facebook ID matched an orphan store record from earlier failed attempts, while the real business store had no Facebook IDs set.
- **Using relative image paths for Meta publishing**: Meta's servers make server-side fetches of image URLs. `http://localhost:3000/demo-products/abaya.png` is unreachable from Meta's infrastructure.

## Solution

### 1. Facebook Login for Business with `config_id`

Create a Configuration in Meta Developer Console (Facebook Login for Business > Configurations) with these permissions: `pages_show_list`, `pages_manage_posts`, `pages_read_engagement`, `instagram_basic`, `instagram_content_publish`. Pass the resulting `config_id` in the OAuth URL instead of `scope`:

```typescript
// app/api/auth/meta/facebook/route.ts
const authUrl = new URL(`https://www.facebook.com/v25.0/dialog/oauth`);
authUrl.searchParams.set("client_id", META_APP_ID);
authUrl.searchParams.set("redirect_uri", redirectUri);
authUrl.searchParams.set("config_id", META_FB_CONFIG_ID); // e.g. "992927346546138"
authUrl.searchParams.set("state", state);
authUrl.searchParams.set("response_type", "code");
authUrl.searchParams.set("auth_type", "rerequest");
```

### 2. Page fetch fallback for dev mode

When `/me/accounts` returns empty, fetch the page directly using `META_PAGE_ID` from env:

```typescript
// In the OAuth callback
if (pages.length === 0) {
  const envPageId = process.env.META_PAGE_ID;
  if (envPageId) {
    const directRes = await fetch(
      `${GRAPH_API_BASE}/${envPageId}?fields=id,name,access_token&access_token=${userAccessToken}`
    );
    if (directRes.ok) {
      const directPage = await directRes.json();
      if (directPage.id) pages = [directPage];
    }
  }
}
```

### 3. Instagram Business account linking via Page token

After fetching the page, query for the linked Instagram Business account:

```typescript
const igRes = await fetch(
  `${GRAPH_API_BASE}/${page.id}?fields=instagram_business_account{id,username}&access_token=${page.access_token}`
);
if (igRes.ok) {
  const igData = await igRes.json();
  if (igData.instagram_business_account) {
    igAccountId = igData.instagram_business_account.id;
    igUsername = igData.instagram_business_account.username;
  }
}
```

### 4. Cookie-based token persistence

Store Meta tokens in both in-memory Map and a base64-encoded httpOnly cookie:

```typescript
// lib/session.ts
export function encodeMetaTokensForCookie(metaTokens: MetaTokens): string {
  return Buffer.from(JSON.stringify(metaTokens)).toString("base64");
}

export function getMetaTokensFromCookie(cookieValue: string | undefined): MetaTokens | null {
  if (!cookieValue) return null;
  try {
    return JSON.parse(Buffer.from(cookieValue, "base64").toString("utf-8"));
  } catch { return null; }
}

// In API endpoints — check memory first, then cookie fallback
let metaTokens = token ? getMetaTokens(token) : null;
if (!metaTokens) {
  metaTokens = getMetaTokensFromCookie(cookieStore.get("meta_tokens")?.value);
}
```

### 5. Store matching by Page ID, then slug

Match OAuth logins to existing stores using the business identity (Page ID) first:

```typescript
// convex/stores.ts — findOrCreateByFacebook
// Priority 1: Match by facebookPageId (stable business identity)
if (args.facebookPageId) {
  const byPageId = allStores.find((s) => s.facebookPageId === args.facebookPageId);
  if (byPageId) { /* link and return */ }
}
// Priority 2: Match by slug derived from page name
const bySlug = await ctx.db.query("stores")
  .withIndex("by_slug", (q) => q.eq("slug", slug)).unique();
if (bySlug) { /* link and return */ }
// Priority 3: Create new store only if no match found
```

### 6. Convex Storage for public image URLs

Upload product images to Convex Storage, which provides publicly accessible URLs:

```typescript
// convex/storage.ts
export const generateUploadUrl = mutation({
  handler: async (ctx) => await ctx.storage.generateUploadUrl(),
});

// Admin endpoint reads local files and uploads to Convex Storage
const uploadUrl = await convexServer.mutation(api.storage.generateUploadUrl, {});
const uploadRes = await fetch(uploadUrl, {
  method: "POST",
  headers: { "Content-Type": contentType },
  body: fileBuffer,
});
const { storageId } = await uploadRes.json();
const publicUrl = await convexServer.query(api.storage.getUrl, { storageId });
// Result: https://greedy-dinosaur-146.convex.cloud/api/storage/...
```

## Why This Works

- **`config_id`** tells Facebook Login for Business which permissions to request and which asset types (Pages, Instagram accounts) to show in the consent screen. Without it, the OAuth flow only grants basic user permissions.
- **Direct page fetch fallback** works because even when `/me/accounts` is empty in dev mode, a direct `/{page_id}` call with a user token that has page permissions succeeds.
- **The `instagram_business_account` field** on a Page is accessible when the token has `instagram_basic` permission — the same page token works for both Facebook and Instagram publishing.
- **httpOnly cookies** persist in the browser regardless of server-side module state. They survive Next.js hot reloads and provide a zero-infrastructure persistence layer.
- **Page ID matching** uses the stable business identity rather than the personal user ID, ensuring OAuth reconnects to the correct store even when one person manages multiple pages.
- **Convex Storage** generates permanent public URLs that Meta's servers can fetch server-side for photo posts and Instagram container creation.

## Prevention

- Always use `config_id` (not `scope`) when working with Facebook Login for Business. Document the required Meta Developer Console configuration steps for new developers.
- Never rely solely on in-memory state for critical auth tokens in Next.js dev. Use cookies or a database as the source of truth.
- Design identity-linking mutations to match on the most specific business identifier first (Page ID), not personal user IDs.
- When integrating with external APIs that fetch media server-side (Meta, Twitter, Slack), ensure images are hosted on publicly accessible URLs from the start. Convex Storage is a good zero-config option.
- Add `instagram_basic` and `instagram_content_publish` to the Facebook Login for Business configuration upfront — don't assume they'll come from a separate Instagram OAuth flow.

## Related Issues

- [docs/solutions/conventions/meta-api-hackathon-workaround-2026-05-20.md](../conventions/meta-api-hackathon-workaround-2026-05-20.md) — Development mode setup, manual token acquisition, and permission prerequisites
- [docs/solutions/architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md](../architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md) — The Next.js + Convex split-responsibility pattern that Meta routes follow
- [docs/solutions/architecture-patterns/multi-tenant-session-auth-convex-2026-05-21.md](../architecture-patterns/multi-tenant-session-auth-convex-2026-05-21.md) — Multi-tenant session-based authorization that hardens the session persistence pattern from this doc into per-store data isolation
