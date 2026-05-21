---
title: "Multi-Tenant Session-Based Authorization for Convex + Next.js"
date: 2026-05-21
category: architecture-patterns
module: authentication
problem_type: architecture_pattern
component: authentication
severity: critical
applies_when:
  - "Multiple sellers share the same Convex deployment and each seller's data must be isolated"
  - "Convex Auth is not available or authentication is handled externally via OAuth"
  - "Any mutation or query exposes data scoped to a tenant (store, organization, team)"
  - "Resources have ownership hierarchies (store -> product -> variant)"
tags:
  - multi-tenancy
  - convex
  - session-management
  - authorization
  - data-isolation
  - nextjs
  - hackathon
---

# Multi-Tenant Session-Based Authorization for Convex + Next.js

## Context

The Swipe Social Storefront is a multi-tenant social commerce platform where multiple independent sellers each have their own store. Built on Convex + Next.js, the app originally had no authentication or authorization on seller-facing Convex functions. Every mutation -- creating products, viewing orders, updating store settings -- accepted a raw `storeId` argument from the client with no verification. Any client could pass any `storeId` and read or modify another seller's data.

This became a blocking issue for the hackathon demo where multiple sellers would sign up via Meta Business OAuth and use the same deployment simultaneously. Without data isolation, Seller A could see Seller B's orders, modify their products, change their store settings, or even read their Swipe payment credentials.

The product vision (auto memory [claude]) is a seller-centric social commerce platform for the Maldives where sellers onboard, manage inventory, and receive payments via Swipe -- all requiring strict per-store data boundaries.

## Guidance

The implementation follows a three-layer pattern: **session creation at login, session validation in every mutation, and ownership checks on nested resources**.

### Layer 1: Persistent sessions table in Convex

Replace in-memory sessions (which are lost on server restart) with a Convex table:

```ts
// convex/schema.ts
sessions: defineTable({
  token: v.string(),
  storeId: v.id("stores"),
  createdAt: v.number(),
  expiresAt: v.number(),
}).index("by_token", ["token"]),
```

### Layer 2: Auth helper -- `authenticateStore()`

A single function that every seller-facing function calls first. It looks up the session by token, checks expiry, and returns the verified `storeId`:

```ts
// convex/auth.ts
export async function authenticateStore(
  ctx: QueryCtx | MutationCtx,
  sessionToken: string,
): Promise<Id<"stores">> {
  const session = await ctx.db
    .query("sessions")
    .withIndex("by_token", (q) => q.eq("token", sessionToken))
    .unique();

  if (!session) throw new Error("Invalid session");
  if (Date.now() > session.expiresAt) throw new Error("Session expired");

  return session.storeId;
}
```

Ownership validators for nested resources compare the resource's `storeId` against the authenticated store:

```ts
export async function validateProductOwnership(
  ctx: QueryCtx | MutationCtx,
  productId: Id<"products">,
  storeId: Id<"stores">,
): Promise<void> {
  const product = await ctx.db.get(productId);
  if (!product || product.storeId !== storeId) {
    throw new Error("Product not found or access denied");
  }
}
```

For deeply nested resources (variant -> product -> store), traverse the chain:

```ts
export async function validateVariantOwnership(
  ctx: QueryCtx | MutationCtx,
  variantId: Id<"productVariants">,
  storeId: Id<"stores">,
): Promise<void> {
  const variant = await ctx.db.get(variantId);
  if (!variant) throw new Error("Variant not found");
  const product = await ctx.db.get(variant.productId);
  if (!product || product.storeId !== storeId) {
    throw new Error("Variant not found or access denied");
  }
}
```

### Layer 3: Session lifecycle

**Creation** -- During OAuth callback, create a Convex session and set both an httpOnly cookie and a client-readable cookie:

```ts
// In Facebook OAuth callback
const sessionToken = createSession(store.storeId);
await convexServer.mutation(api.sessions.create, {
  token: sessionToken,
  storeId: store.storeId,
});
// Set client-readable cookie so React can pass token to Convex
response.cookies.set("seller_session_token", sessionToken, {
  httpOnly: false, sameSite: "lax", path: "/",
});
```

**Client access** -- A React hook reads the token from cookies:

```ts
// lib/use-session.ts
export function useSessionToken(): string | null {
  const [token, setToken] = useState<string | null>(null);
  useEffect(() => {
    const match = document.cookie.split(";")
      .map((c) => c.trim())
      .find((c) => c.startsWith("seller_session_token="));
    if (match) setToken(decodeURIComponent(match.split("=")[1]));
  }, []);
  return token;
}
```

### Applying auth to every seller-facing function

Every seller-facing function adds `sessionToken: v.string()` to its args and calls `authenticateStore()` as its first operation. The returned `storeId` replaces any client-supplied storeId.

## Why This Matters

Without this pattern:

- **Data leakage**: Any seller can read another seller's orders, revenue stats, and customer data
- **Data corruption**: Any seller can modify another seller's products, fulfill their orders, or change store settings
- **Credential theft**: Store-level secrets (Swipe payment API keys) could be read or overwritten
- **Demo failure**: Multiple test sellers on the same deployment see each other's data

Even in single-tenant prototypes, adding this pattern early prevents a costly retrofit later.

## When to Apply

- Multiple users share the same Convex deployment and each user's data must be isolated
- Convex Auth is not available or authentication is handled externally via OAuth
- Any mutation or query exposes data scoped to a tenant (store, organization, team)
- Resources have ownership hierarchies (store -> product -> variant)

Do NOT apply to genuinely public endpoints (storefront browsing, order status by access token) -- those should remain unauthenticated by design.

## Examples

### Before (no auth -- trusts client-supplied storeId):

```ts
// Any caller can list any store's orders
export const listByStore = query({
  args: { storeId: v.id("stores") },
  handler: async (ctx, args) => {
    return await ctx.db.query("orders")
      .withIndex("by_storeId", (q) => q.eq("storeId", args.storeId))
      .take(200);
  },
});
```

### After (session-authenticated -- storeId derived server-side):

```ts
export const listByStore = query({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    return await ctx.db.query("orders")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .order("desc")
      .take(200);
  },
});
```

### Client-side -- passing session token:

```tsx
const sessionToken = useSessionToken();
const stats = useQuery(
  api.orders.getDashboardStats,
  sessionToken ? { sessionToken } : "skip"
);
```

The `"skip"` sentinel prevents the query from firing before the token is available, avoiding auth errors during initial render.

## Related

- [Swipe/Next.js/Convex Split-Responsibility Architecture](../../../docs/solutions/architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md) -- architectural context for why Convex mutations need their own auth layer
- [Meta Business OAuth & Social Publishing](../integration-issues/meta-business-oauth-social-publishing-2026-05-21.md) -- complementary session persistence pattern for Meta tokens
- [Original implementation plan](../../../docs/plans/2026-05-20-001-feat-swipe-social-storefront-plan.md) -- original session design that was hardened
