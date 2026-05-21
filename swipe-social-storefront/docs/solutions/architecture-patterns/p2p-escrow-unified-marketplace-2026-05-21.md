---
title: "P2P USDT Exchange with Simulated Escrow and Unified Marketplace"
date: 2026-05-21
category: architecture-patterns
module: exchange
problem_type: architecture_pattern
component: backend
severity: medium
applies_when:
  - "Building a P2P exchange with escrow on a serverless database (Convex)"
  - "Simulating blockchain operations for hackathon or MVP"
  - "Adding a second product pillar to an existing marketplace app"
  - "Handling concurrent partial purchases on a shared inventory pool"
tags:
  - p2p-exchange
  - escrow
  - concurrency-control
  - convex
  - nextjs
  - crypto
  - marketplace
  - reservation-pattern
  - navigation-shells
---

# P2P USDT Exchange with Simulated Escrow and Unified Marketplace

## Context

A P2P USDT-to-fiat (MVR) crypto exchange needed to be built as a second product pillar alongside an existing social commerce marketplace, both sharing the same Convex database and Next.js app. The target market (Maldives) has limited crypto on-ramp options, and the hackathon timeline demanded simulated blockchain operations while maintaining production-realistic data flows. The core challenge: how do you build escrow, concurrency-safe partial purchases, and multi-tenant payment routing without a real blockchain or custodial wallet service?

Key constraints:
- All crypto operations simulated in Convex (no blockchain)
- Swipe (local payment gateway) is the only real payment rail
- Sellers may be marketplace-only, crypto-only, or both
- Buyers are unauthenticated — zero friction, details at checkout only
- Must demonstrate the concept convincingly to hackathon judges

## Guidance

### 1. Simulated Escrow via Convex Mutations

All crypto operations are simulated entirely within Convex's transactional mutation layer. No external blockchain calls occur.

- **Wallet generation**: `createListing` generates a TRC20-format address (`"T" + randomAlphanumeric(33)`) stored on the listing record. This address is cosmetic — it gives sellers a realistic deposit UX without chain integration.
- **Deposit confirmation**: `confirmDeposit` flips `status` from `"pending_deposit"` to `"active"` and sets `availableBalance = usdtAmount`. In production, this would be triggered by a blockchain watcher; for hackathon, it's a button click.
- **Transaction hash**: `confirmPurchase` generates `"0x" + randomHex(64)` as a simulated tx hash stored on the transaction record.

The data model is identical to what a real implementation would use — only the wallet generation and tx hash functions change in production.

### 2. Reservation Model for Concurrent Partial Purchases

The classic P2P exchange race condition: two buyers try to purchase from the same listing simultaneously. Solved with soft reservations and a 5-minute TTL:

```typescript
// createReservation mutation (atomic within Convex)
// 1. Check available balance
if (args.amount > listing.availableBalance) throw new Error("Insufficient balance");
// 2. Atomic decrement/increment
await ctx.db.patch(args.listingId, {
  availableBalance: listing.availableBalance - args.amount,
  reservedBalance: listing.reservedBalance + args.amount,
});
// 3. Create reservation with 5-min TTL
await ctx.db.insert("exchangeReservations", {
  listingId, amount, buyerWallet,
  expiresAt: Date.now() + 5 * 60 * 1000,
  status: "active",
});
```

- **On payment confirmation**: reservation marked `"completed"`, `reservedBalance` decremented. If both balances hit zero, listing becomes `"sold_out"`.
- **On expiry**: `expireReservation` (internal mutation via scheduler) restores amount to `availableBalance`.
- **Fallback guard**: `getActiveListings` query filters `availableBalance <= 0` at query time, catching any scheduler delays.

Key insight: Convex mutations are serialized per-document, so the read-modify-write on `availableBalance`/`reservedBalance` within a single mutation is inherently atomic — no explicit locks needed.

### 3. Platform-Level Payment Credentials for Exchange

Exchange purchases use platform-level Swipe credentials (environment fallback), not per-seller credentials. Crypto-only sellers never need to configure Swipe merchant accounts — the platform collects MVR on their behalf.

```typescript
// Exchange: platform credentials (no seller setup)
const swipePayment = await createPayment(mvrAmount, "MVR", description);
// Uses SWIPE_CLIENT_ID / SWIPE_CLIENT_SECRET from env

// Marketplace: per-seller credentials (seller must onboard)
const swipePayment = await createPaymentWithCredentials(
  amount, currency, desc, store.swipeClientId, store.swipeClientSecret
);
```

### 4. Multi-Path Onboarding via sellerType

A `sellerType` field on the `stores` table (`"marketplace"` | `"crypto"` | `"both"`) drives conditional rendering throughout:
- Dashboard sections (marketplace stats, crypto stats, or both)
- Navigation items (Exchange tab shown only for crypto/both sellers)
- Onboarding steps (Swipe creds for marketplace, wallet address for crypto)

### 5. Route-Based Shell Dispatch for Multi-Persona Apps

Buyer-facing routes (`/marketplace`, `/exchange`) use a distinct `BuyerShell` with its own navigation, completely separate from the seller shell:

```typescript
// layout-router.tsx
const isBuyer = pathname.startsWith('/marketplace') || pathname.startsWith('/exchange');
const isLanding = pathname === '/';
if (isLanding) return <>{children}</>;  // No shell
if (isBuyer) return <BuyerShell>{children}</BuyerShell>;
// Falls through to seller shell
```

### 6. Schema: Three Exchange Tables

```
exchangeListings:     storeId, usdtAmount, rate, availableBalance, reservedBalance,
                      partialAllowed, walletAddress, status, createdAt
exchangeTransactions: listingId, buyerWallet, usdtAmount, mvrAmount, txHash,
                      status, swipePaymentId, createdAt
exchangeReservations: listingId, amount, buyerWallet, expiresAt, status
```

Indexes: `by_status` and `by_storeId` on listings; `by_listingId` on transactions and reservations.

## Why This Matters

- **Simulated escrow** lets you demo a production-realistic flow without blockchain dependencies. The data model translates directly to production — swap random generators for real API calls.
- **The reservation pattern** solves the overselling race condition that would break trust in a P2P exchange. The 5-minute TTL prevents capital lockup from abandoned purchases.
- **Platform-level credentials** remove the biggest friction point for crypto sellers — they can list USDT without any payment gateway setup.
- **Route-based shell dispatch** keeps buyer and seller experiences independent without complex role/auth checks. Adding a new product vertical is just a pathname prefix and a shell component.

## When to Apply

- **Simulated escrow**: Any hackathon/MVP with blockchain, payment escrow, or custodial operations where business logic must be correct but external integrations are unavailable. Replace random generators with real API calls for production.
- **Reservation with TTL**: Whenever you have a shared inventory pool with concurrent buyers and partial consumption — ticket sales, limited-stock flash sales, or any subdivided resource. Critical when payment takes user time (redirect to payment page).
- **Platform-level credentials**: When onboarding multiple seller types with different payment needs, and one vertical shouldn't require the same merchant setup as another.
- **Route-prefix shell dispatch**: When a single Next.js app serves multiple user personas (buyer/seller/admin) with distinct navigation and layout patterns.

## Examples

**Key files in this implementation:**

| File | Purpose |
|------|---------|
| `convex/exchange.ts` | All escrow, reservation, and transaction mutations |
| `convex/schema.ts` | Exchange table definitions with indexes |
| `app/api/exchange/purchase/route.ts` | Purchase initiation with platform credentials |
| `app/api/exchange/purchase/[purchaseId]/status/route.ts` | Payment polling + escrow release |
| `components/shells/layout-router.tsx` | Route-based shell dispatch |
| `components/shells/buyer-shell.tsx` | Buyer navigation shell |
| `app/seller/onboarding/page.tsx` | Multi-path onboarding with sellerType |

**Parallel implementation strategy**: 10 implementation units dispatched as independent background agents with no file overlap. Schema first, then backend mutations, then UI in parallel. 4,667 lines implemented in approximately 15 minutes wall clock via subagent parallelism.

## Related

- `docs/solutions/architecture-patterns/multi-tenant-session-auth-convex-2026-05-21.md` — The session auth pattern every exchange mutation uses
- `docs/solutions/integration-issues/meta-business-oauth-social-publishing-2026-05-21.md` — Upstream OAuth flow feeding into exchange onboarding (may need `sellerType` field update)
- `docs/brainstorms/2026-05-21-swipe-crypto-exchange-marketplace-requirements.md` — Origin requirements document
- `docs/plans/2026-05-21-001-feat-crypto-exchange-marketplace-plan.md` — Implementation plan
