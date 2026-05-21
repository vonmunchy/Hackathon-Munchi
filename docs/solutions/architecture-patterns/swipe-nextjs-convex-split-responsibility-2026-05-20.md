---
title: "Swipe Payment Integration: Next.js + Convex Split-Responsibility Architecture"
date: 2026-05-20
category: architecture-patterns
module: swipe-social-storefront
problem_type: architecture_pattern
component: payments
severity: medium
applies_when:
  - "Integrating a payment API (or any external service) that runs on localhost or a private network with a Convex backend"
  - "Using Convex actions that cannot reach localhost services due to cloud execution"
  - "SSE or streaming endpoints require server-side OAuth credentials the browser cannot hold"
  - "Concurrent inventory purchases require atomic stock deduction without reservation logic"
  - "Real-time seller dashboard updates needed without WebSocket infrastructure"
tags:
  - swipe-api
  - nextjs
  - convex
  - sse-proxy
  - atomic-transactions
  - real-time
  - payments
  - hackathon
---

# Swipe Payment Integration: Next.js + Convex Split-Responsibility Architecture

## Context

This guidance emerged from building "Swipe Social Storefront" — a hackathon MVP enabling Maldivian Instagram/Facebook sellers to accept BML Swipe payments through a Next.js App Router + Convex + Tailwind stack.

The core challenge: Swipe's payment API runs as a local mock server at `localhost:8080` during development (and behind a private network in production), while **Convex actions execute in Convex's own cloud infrastructure**. This creates a hard network boundary — Convex cloud cannot reach `localhost` — that shapes the entire application architecture.

Additional constraints discovered during a 7-persona document review:
- The Swipe SSE endpoint (`/api/v1/payments/{id}/stream`) requires an OAuth2 Bearer token — browsers cannot set Authorization headers on `EventSource` connections
- The Swipe mock's "simulate payment" endpoint uses `POST /pay/{shortCode}/complete` (not `/pay/{reference}/complete`) and returns a 303 redirect, not JSON
- The mock auto-completes PENDING payments after 60 seconds (`PaymentTransitionTTL`)
- Concurrent purchases of the last inventory unit must be handled atomically (no reservation system — stock deducts only on payment confirmation)

## Guidance

### 1. Split Responsibility: Next.js = Network Bridge, Convex = State Machine

Divide system responsibilities based on what each runtime can reach:

**Next.js API Routes own all external API communication:**
- OAuth2 client credentials flow (Basic Auth to `/oauth2/token`, cache Bearer token in server memory)
- Payment creation (`POST /api/v1/payments` with type: QR)
- SSE proxy (authenticate to Swipe SSE, re-emit events to browser)
- Payment simulation (`POST /pay/{shortCode}/complete`)
- Wallet balance and transaction history (`GET /api/v1/balance`, `GET /api/v1/history`)

**Convex owns all data persistence and real-time state:**
- Product catalog (stock levels, pricing, variants, metadata)
- Order records and payment status
- Atomic stock deductions via serializable mutations
- Real-time dashboard subscriptions via `useQuery`
- Seed data management

**The bridge pattern:** Next.js API routes call the Swipe API, then call Convex mutations to persist the results. The browser never connects to Swipe directly.

### 2. SSE Proxy Pattern

The browser cannot connect directly to Swipe's SSE endpoint because:
1. SSE requires an OAuth Bearer token — the `EventSource` API has no way to set `Authorization` headers
2. The Swipe mock checks merchant ownership (`auth.PrincipalFromContext`) and returns 401/403 without valid credentials
3. CORS would block cross-origin requests from `localhost:3000` to `localhost:8080`

**Solution:** A Next.js API route at `/api/checkout/[orderId]/stream` that:
1. Looks up the order in Convex to get the `swipePaymentId`
2. Retrieves the cached merchant Bearer token
3. Opens a server-side fetch to Swipe's SSE endpoint with the Bearer token
4. Creates a `ReadableStream` that re-emits each SSE event to the browser
5. Returns a `text/event-stream` response — the browser connects via `EventSource` to the Next.js route (same origin, no CORS)
6. On terminal status (COMPLETED/EXPIRED/CANCELLED), calls a Convex mutation to update the order, then closes the stream

**Fallback:** If SSE fails (connection error, 3 retries exhausted), the browser falls back to polling a Next.js API route every 5 seconds that calls `GET /api/v1/payments/{id}`.

### 3. Atomic Stock Deduction via Convex Serializable Transactions

Do not use stock reservation (reserve-on-intent, release-on-timeout). Instead, use atomic compare-and-decrement on payment confirmation:

```
Convex mutation: confirmPayment(orderId)
  1. Get order — verify paymentStatus !== "paid" (idempotent)
  2. Get variant — read current stockAvailable
  3. If stockAvailable < order.quantity:
       → Patch order to cancelled, return insufficient_stock
  4. If stockAvailable >= order.quantity:
       → Patch variant: stockAvailable -= quantity, stockSold += quantity
       → Patch order: status = "paid", paymentStatus = "paid"
       → Return success
```

**Why this works:** All reads and writes inside a single Convex mutation run in a serializable transaction. Two concurrent calls for the last unit will serialize automatically — the first succeeds and deducts stock, the second reads the updated stock and fails with "insufficient stock." No distributed lock, no reservation table, no cron cleanup needed.

### 4. Payment Display: QR + Link on All Viewports

Always create `QR`-type payment intents (the Swipe mock guarantees support). From a single QR payment, extract both:
- `qr_data` (base64 PNG) — render as `<img>` for scanning
- Mock pay page URL at `/pay/{shortCode}` — render as tappable button/link

Show both on all viewports:
- Desktop: QR prominently displayed, link as secondary
- Mobile: Link button prominently displayed, smaller QR below

Never omit one based on viewport detection. The same payment object carries both artifacts.

### 5. Demo Mode (SWIPE_DEMO_MODE)

Environment variable `SWIPE_DEMO_MODE=true` causes all Next.js Swipe API routes to return hardcoded mock responses:
- `createPayment()` returns a demo payment with placeholder QR
- `streamPaymentStatus()` simulates SSE: emits PENDING immediately, COMPLETED after 3 seconds
- `getWalletBalance()` and `getTransactionHistory()` return static demo data
- `simulatePayment()` transitions a demo payment to COMPLETED in memory

This decouples the demo from the Swipe mock process entirely. Convex is unaware of demo mode — it always receives real mutation calls.

### 6. Convex Real-Time for Seller Dashboard

Use Convex `useQuery` in seller dashboard components. When a buyer's payment completes and the Next.js API route calls a Convex mutation to confirm the order, the seller's dashboard auto-updates — order counts, stock levels, and sales totals change without polling, manual refresh, or WebSocket setup.

This natively satisfies the "new order indicator" requirement without any additional infrastructure.

## Why This Matters

| If you... | What breaks |
|-----------|------------|
| Call Swipe from Convex actions | Connection refused — Convex cloud cannot reach `localhost:8080` |
| Let browser connect to Swipe SSE directly | 401 Unauthorized — no Bearer token on `EventSource` |
| Use stock reservation instead of atomic decrement | Unnecessary complexity: reservation expiry cron, abandoned checkout release, late-payment edge cases |
| Create LINK-type payments (unverified mock support) | Payment creation may fail at demo time |
| Only show QR on mobile | Buyer can't scan their own screen |
| Only show link on desktop | Lose scan-to-pay convenience for counter/kiosk scenarios |
| Skip demo mode | Every dev session and live demo requires a running Swipe mock process |
| Poll for dashboard updates instead of using Convex real-time | Unnecessary latency, complexity, and network overhead |

## When to Apply

This pattern applies when **all** of the following are true:
- Backend uses **Convex** (or any serverless/BaaS that runs in a remote cloud and cannot reach localhost)
- Third-party API runs on **localhost in dev** or on a **private/internal network**
- The third-party API uses **SSE or streaming** with server-side authentication
- You need **concurrent-safe inventory operations** (e-commerce, ticketing, booking)
- You need **real-time UI updates** without building WebSocket infrastructure
- You need a **demo mode** that works offline from the third-party service

The pattern generalizes beyond Swipe/Maldives to any stack where:
- A BaaS (Convex, Firebase Functions, Supabase Edge Functions) is the persistence layer
- A payment gateway or external API is network-isolated from that BaaS
- The frontend framework (Next.js, Nuxt, SvelteKit) can serve API routes from the same origin

## Examples

### Architecture Diagram

```
Browser (Buyer)                    Browser (Seller)
     │                                  │
     ├─ fetch /api/swipe/*              ├─ useQuery (Convex real-time)
     ├─ EventSource /api/checkout/      │
     │   [orderId]/stream               │
     │                                  │
     └────────────┬─────────────────────┘
                  │
          Next.js API Routes
                  │
        ┌─────────┼──────────┐
        │         │          │
   Swipe Client   │    Demo Mode
   (OAuth+HTTP)   │    (mock responses)
        │         │
        ▼         ▼
   Swipe Mock   Convex Cloud
   :8080        (DB + real-time
                 + atomic mutations)
```

### Payment Creation Flow

```
1. Buyer clicks "Buy Now" → POST /api/swipe/payments/create
2. Next.js API route:
   a. Look up product/variant price in Convex (server-side — R35 price validation)
   b. Create order in Convex (status: pending)
   c. Get/refresh OAuth Bearer token from Swipe
   d. POST /api/v1/payments {amount, currency: "MVR", type: "QR", description}
   e. Update order in Convex with swipePaymentId, swipeShortCode
   f. Return {orderId, qrData, shortCode} to browser
3. Browser renders QR image + pay-page link button
4. Browser opens EventSource to /api/checkout/[orderId]/stream
5. Next.js proxies Swipe SSE with Bearer token
6. On COMPLETED: call Convex confirmPayment mutation → atomic stock deduction
7. Browser redirects to success page
```

### Convex Atomic Mutation (Directional)

```typescript
// convex/orders.ts — confirmPayment
export const confirmPayment = mutation({
  args: { orderId: v.id("orders") },
  handler: async (ctx, { orderId }) => {
    const order = await ctx.db.get(orderId);
    if (!order) throw new ConvexError("order_not_found");
    if (order.paymentStatus === "paid") return { success: true }; // idempotent

    const variant = await ctx.db.get(order.variantId);
    if (!variant || variant.stockAvailable < order.quantity) {
      await ctx.db.patch(orderId, { status: "cancelled", paymentStatus: "cancelled" });
      return { success: false, reason: "insufficient_stock" };
    }

    // Atomic: Convex serializable transaction
    await ctx.db.patch(variant._id, {
      stockAvailable: variant.stockAvailable - order.quantity,
      stockSold: variant.stockSold + order.quantity,
    });
    await ctx.db.patch(orderId, {
      status: "paid",
      paymentStatus: "paid",
      updatedAt: Date.now(),
    });
    return { success: true };
  },
});
```

### Next.js SSE Proxy Route (Directional)

```typescript
// app/api/checkout/[orderId]/stream/route.ts
export async function GET(request, { params }) {
  const { orderId } = params;
  // Look up order to get swipePaymentId...
  const token = await getSwipeAccessToken(); // cached in-memory

  if (process.env.SWIPE_DEMO_MODE === "true") {
    // Simulate: PENDING → 3s delay → COMPLETED
    return new Response(demoSSEStream(), {
      headers: { "Content-Type": "text/event-stream" },
    });
  }

  const upstream = await fetch(
    `${SWIPE_BASE_URL}/api/v1/payments/${paymentId}/stream`,
    { headers: { Authorization: `Bearer ${token}` } }
  );

  return new Response(upstream.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    },
  });
}
```

## Related

- **Origin requirements:** `docs/brainstorms/2026-05-20-swipe-social-storefront-requirements.md`
- **Requirements review changelog:** `docs/brainstorms/2026-05-20-requirements-review-changelog.md`
- **Implementation plan:** `docs/plans/2026-05-20-001-feat-swipe-social-storefront-plan.md`
- **Swipe API spec:** `swipe-merchants-dev/spec/app.yaml`
- **Swipe reference integration:** `swipe-merchants-dev/example/main.go`
- **Convex + Next.js docs:** https://docs.convex.dev/client/nextjs/app-router/
- **Convex mutations (atomic transactions):** https://docs.convex.dev/functions/mutation-functions
- **Convex actions (external HTTP calls):** https://docs.convex.dev/functions/actions
- **Multi-tenant session auth:** `swipe-social-storefront/docs/solutions/architecture-patterns/multi-tenant-session-auth-convex-2026-05-21.md` — adds per-store authorization to the Convex mutation layer described in this doc
