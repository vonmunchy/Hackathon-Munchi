---
title: "Meta API App Review Takes 2-6 Weeks: Development Mode Workaround for Hackathons"
date: 2026-05-20
last_updated: 2026-05-21
category: conventions
module: meta-integration
problem_type: convention
component: payments
severity: high
applies_when:
  - "Building an app that needs Facebook Page or Instagram account access for a hackathon or time-boxed demo"
  - "Sellers need to connect their social media pages but Meta App Review takes 2-6 weeks"
  - "You need to demo real Facebook/Instagram API integration without production approval"
  - "Your app reads Page posts, manages Page messaging, or accesses Instagram Business accounts"
  - "pages_manage_posts permission is missing from Graph API Explorer dropdown"
  - "GET /me/accounts returns empty data array despite being a Page admin"
tags:
  - meta-api
  - facebook-pages
  - instagram
  - hackathon
  - app-review
  - development-mode
  - oauth
  - social-commerce
  - graph-api-v25
  - use-case-permissions
  - page-access-token
  - long-lived-token
---

# Meta API App Review Takes 2-6 Weeks: Development Mode Workaround for Hackathons

## Context

When building Swipe Social Storefront — a social commerce platform for Maldives Instagram/Facebook sellers — we needed sellers to connect their Facebook Pages and Instagram accounts so the app could read their product posts, import product data, and facilitate DM-based payment flows.

The problem: accessing a user's Facebook Page or Instagram Business account via the Meta Graph API requires **Meta App Review**, which takes **2-6 weeks**. For a hackathon with a days-long timeline, this is a hard blocker for the standard production path.

We researched the actual Meta developer documentation and found a legitimate alternative that provides **real API access to real Facebook Pages** without waiting for App Review.

## Guidance

### The Two-Tier Approach

**Tier 1 — Hackathon Demo (works immediately, no approval):**

Use Meta's **Development Mode** with your own accounts. The key finding from Meta's documentation:

> "Apps in Development Mode can request any Permission from any app User who has a Role on the app."

> "Individuals with Administrator, Developer, or Tester roles can grant the app any permission while it is in development without requiring app review."

This means: if you (the developer) are also the Facebook Page owner AND have an Admin/Developer role on your Meta App, you can grant yourself **all permissions** — including `pages_read_engagement`, `pages_manage_posts`, `pages_messaging`, `instagram_basic` — and access your **real Facebook Page and Instagram account** in development mode.

**Tier 2 — Production (requires Meta App Review):**

When other sellers (not your team) need to connect their Pages, submit for Meta App Review after the hackathon.

### What Each Permission Level Requires

| Permission | Access Level | App Review? | Works in Dev Mode? |
|-----------|-------------|-------------|-------------------|
| `public_profile` | Standard | No | Yes |
| `pages_show_list` | Standard | No | Yes |
| `pages_read_engagement` | Advanced | Yes (production) | **Yes (dev mode for app role holders)** |
| `pages_manage_posts` | Advanced | Yes (production) | **Yes (dev mode for app role holders)** |
| `pages_messaging` | Advanced | Yes (production) | **Yes (dev mode for app role holders)** |
| `instagram_basic` | Advanced | Yes (production) | **Yes (dev mode for app role holders)** |
| `instagram_content_publish` | Advanced | Yes (production) | **Yes (dev mode for app role holders)** |

### Who Can Use the App in Development Mode

| Person | Can use? | How |
|--------|----------|-----|
| You (app Admin) | Yes | Full access to your own Pages |
| Teammate (added as Developer/Tester) | Yes | Full access to their own Pages |
| Up to ~50 people added as Testers | Yes | Full access to their own Pages |
| Random real seller (no role on app) | **No** | Needs App Review + Live Mode |

### Step-by-Step Setup (Hackathon)

1. **Create a real Facebook Page** (e.g., "Island Finds MV") from your personal account. Post 3-5 product photos.
2. **Create a real Instagram Business account** linked to the Facebook Page. Post the same products.
3. **Create a Meta Developer App** at developers.facebook.com (instant, free). Choose "Business" type.
4. **Add Facebook Login for Business** as a product in the app dashboard.
5. **Configure OAuth redirect**: `http://localhost:3000/api/auth/meta/callback`
6. **Get tokens via Graph API Explorer** (developers.facebook.com/tools/explorer/):
   - Select your app
   - Grant all needed permissions (all work in dev mode for you)
   - Generate User Access Token
   - Call `GET /me/accounts` to get your Page Access Token
7. **Exchange for long-lived token**: `GET /oauth/access_token?grant_type=fb_exchange_token&client_id={APP_ID}&client_secret={APP_SECRET}&fb_exchange_token={SHORT_TOKEN}`
8. The Page Access Token derived from a long-lived User Token **never expires** — use it in your `.env.local`
9. **Test API calls**: `GET /{PAGE_ID}/posts?fields=message,full_picture,created_time`

### What NOT to Use: Test Users

Meta's "Test Users" feature is **not** the right solution for this. Test Users are sandboxed accounts that can only interact with test Pages, not real Facebook Pages. They cannot manage real Pages or access real Instagram accounts. We initially considered this but verified it won't produce a convincing demo with real social media content.

### Sending Payment Links in DMs

A separate concern we verified: **sending third-party payment links (like Swipe) in WhatsApp, Instagram DMs, and Messenger is allowed** by Meta's policies. No policy explicitly prohibits it. Major payment companies (Stripe, Razorpay, PayPal, Square) actively promote sharing payment links via WhatsApp and social DMs. This is the standard conversational commerce pattern used by millions of businesses globally.

The only restriction is: don't spam unsolicited payment links to strangers. In our flow, the buyer initiates contact first, so it's clearly allowed.

## Why This Matters

**If you skip this and try to get App Review for a hackathon:**
- You'll wait 2-6 weeks and have nothing to demo
- Your core payment flow works, but the social integration is a mockup
- Judges see placeholder data instead of real Facebook/Instagram content

**If you use Development Mode correctly:**
- Judges see real Facebook Page posts, real Instagram content, real API integration
- The demo shows actual product import from a real social media presence
- You can truthfully say "the integration works — production access is pending Meta App Review"

**If you use Test Users instead of Development Mode with your own accounts:**
- You get sandboxed test Pages with no real content
- The demo looks fake — judges can tell it's not a real Facebook Page
- You wasted time setting up test accounts when your own account works better

## When to Apply

- Any hackathon or time-boxed project that needs Facebook Page or Instagram API access
- Demos where you want to show real social media data, not mocked content
- Early-stage MVPs before Meta App Review is complete
- Proof-of-concept builds that need to validate the social commerce flow works end-to-end

Does NOT apply when:
- You need random users (not your team) to connect their Pages — that requires App Review
- You're deploying to production with real customers — submit for review
- You only need deep links (WhatsApp `wa.me/` links, Instagram `ig.me/m/` links) — those work without any Meta API or developer app at all

## Examples

### The Pitch Framing for Judges

> "We've built the Facebook Page and Instagram integration using Meta's Graph API. In our demo, we're running in development mode with our own seller Page — you can see real posts, real engagement data, and real product import. For production launch, we'll complete Meta App Review to allow any Maldivian seller to connect their Page. The review submission is ready."

### API Calls That Work in Development Mode

```bash
# Read your Page's posts (real content, real images)
curl "https://graph.facebook.com/v21.0/{PAGE_ID}/posts?fields=message,full_picture,created_time&access_token={PAGE_TOKEN}"

# Read your Page's info
curl "https://graph.facebook.com/v21.0/{PAGE_ID}?fields=name,about,followers_count,picture&access_token={PAGE_TOKEN}"

# Read linked Instagram media
curl "https://graph.facebook.com/v21.0/{IG_ID}/media?fields=caption,media_url,timestamp&access_token={PAGE_TOKEN}"
```

### Environment Variables

```env
META_APP_ID=your_app_id
META_APP_SECRET=your_app_secret
META_PAGE_ID=your_page_id
META_PAGE_ACCESS_TOKEN=your_long_lived_page_token
META_IG_ACCOUNT_ID=your_instagram_business_account_id
```

### The DM-Based Payment Flow (No Meta API Required)

The core seller-buyer flow uses deep links, not the Meta API:

1. Buyer browses storefront, selects product, fills delivery form
2. Buyer taps "Order via WhatsApp" → `https://wa.me/9607771234?text=Hi! I'd like to order: Black Abaya, Size M, MVR 650...`
3. Seller receives WhatsApp message with full order details
4. Seller creates Swipe payment link in dashboard → copies link
5. Seller pastes Swipe payment link back in WhatsApp DM
6. Buyer taps link → pays via Swipe → order confirmed automatically

This entire flow works without any Meta API access, App Review, or developer app. The deep links are just URLs.

## 2025/2026 Use-Case-Based Permission System (Updated)

As of 2025/2026, Meta's App Dashboard has restructured permissions around **"Use cases"**. Permissions are no longer freely selectable in Graph API Explorer — they must first be enabled within a Use Case in the App Dashboard.

### Critical: Adding `pages_manage_posts`

The `pages_manage_posts` permission is **NOT** available in the "Manage messaging & content on Instagram" or "Engage with customers on Messenger from Meta" use cases. You must add a separate use case:

1. Go to App Dashboard → **Use cases** → **"Add use cases"** button
2. Filter by **"Content management"** (not Featured, not Business messaging)
3. Select **"Manage everything on your Page"** — "Publish content and videos, moderate posts and comments"
4. Click **Save**
5. Go back to the new use case → **Customize** → **Permissions and features**
6. Find `pages_manage_posts` and click **"+ Add"**

Without this step, `pages_manage_posts` will not appear in the Graph API Explorer dropdown, and `GET /me/accounts` will return `{ "data": [] }` with no error — a silent failure that wastes hours debugging.

### Token Exchange for Never-Expiring Page Token

Graph API Explorer generates **short-lived tokens** that expire in 1-2 hours. For any durable integration:

```bash
# Step 1: Exchange short-lived user token for long-lived (60-day) user token
GET https://graph.facebook.com/v25.0/oauth/access_token
  ?grant_type=fb_exchange_token
  &client_id={APP_ID}
  &client_secret={APP_SECRET}
  &fb_exchange_token={SHORT_LIVED_USER_TOKEN}

# Step 2: Get never-expiring Page Access Token
GET https://graph.facebook.com/v25.0/{PAGE_ID}
  ?fields=id,name,access_token,instagram_business_account
  &access_token={LONG_LIVED_USER_TOKEN}
```

The Page Access Token returned in Step 2 **never expires** when derived from a long-lived user token. Store it in your `.env.local`.

### Publishing to Facebook Page

```bash
# Text + link post
POST https://graph.facebook.com/v25.0/{PAGE_ID}/feed
  message=Your post content here
  link=https://your-storefront.com/product/123
  access_token={PAGE_ACCESS_TOKEN}

# Photo + caption post
POST https://graph.facebook.com/v25.0/{PAGE_ID}/photos
  url=https://public-image-url.com/product.jpg
  message=Product caption text
  access_token={PAGE_ACCESS_TOKEN}
```

### Silent Failure: Empty `me/accounts`

When `GET /me/accounts` returns `{ "data": [] }`:
- This does **NOT** mean you have no Pages
- It means the token **lacks `pages_manage_posts` or `pages_show_list` permission**
- Verify token permissions at `GET /debug_token?input_token={TOKEN}&access_token={APP_ID}|{APP_SECRET}`
- Fix: ensure the "Manage everything on your Page" use case is added and `pages_manage_posts` is enabled

### Working Next.js Implementation

```typescript
// lib/meta-client.ts
export async function publishToFacebook(message: string, link?: string) {
  const params = new URLSearchParams({
    message,
    access_token: process.env.META_PAGE_ACCESS_TOKEN!,
  });
  if (link) params.set("link", link);

  const res = await fetch(
    `https://graph.facebook.com/v25.0/${process.env.META_PAGE_ID}/feed`,
    { method: "POST", body: params }
  );

  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error?.message || "Publish failed");
  }
  return res.json(); // { id: "pageId_postId" }
}
```

## Related

- **Full OAuth + Publishing implementation**: `swipe-social-storefront/docs/solutions/integration-issues/meta-business-oauth-social-publishing-2026-05-21.md` — Covers the complete Facebook Login for Business OAuth flow with `config_id`, Instagram linking, session persistence via cookies, and Convex Storage for public image URLs. Built on top of the dev-mode workarounds documented here.
- **Architecture pattern**: `docs/solutions/architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md`
- **Implementation plan**: `docs/plans/2026-05-20-001-feat-swipe-social-storefront-plan.md`
- **Requirements**: `docs/brainstorms/2026-05-20-swipe-social-storefront-requirements.md`
- **Meta Graph API docs**: https://developers.facebook.com/docs/pages-api/overview
- **Meta App Roles**: https://developers.facebook.com/docs/development/build-and-test/app-roles/
- **Meta App Modes**: https://developers.facebook.com/docs/development/build-and-test/app-modes/
- **Graph API v25.0 Changelog**: https://developers.facebook.com/docs/graph-api/changelog/version25.0/
- **Meta App Modes**: https://developers.facebook.com/docs/development/build-and-test/app-modes/
