---
title: "feat: UI polish, Binance P2P exchange restyle, admin dashboard, and Swipe demo logging"
type: feat
status: active
date: 2026-05-21
origin: docs/brainstorms/2026-05-21-ui-polish-exchange-admin-requirements.md
---

# UI Polish, Binance P2P Exchange, Admin Dashboard, and Swipe Demo Logging

## Summary

Restyle the exchange browse page from a card grid to a Binance P2P-style order book table (no backend changes needed), build a passphrase-protected admin dashboard for demo seller resets, centralize Swipe API logging in the client library for colored terminal output, and run a visual polish pass across all seller and buyer pages to fix layout issues and apply the amethyst/ruby/slate design system consistently.

---

## Problem Frame

Hackathon judging is imminent. The app works functionally but has visible UI jank, the exchange doesn't evoke the "real crypto platform" feel judges expect, and the presenter has no way to reset sellers or show live payment simulation during the demo. (See origin: `docs/brainstorms/2026-05-21-ui-polish-exchange-admin-requirements.md`)

---

## Requirements

- R1. Fix overlapping elements, z-index conflicts, and layout overflow across all enumerated seller and buyer pages
- R2. Consistent amethyst/ruby/slate design system application (typography, spacing, radius, shadows, palette)
- R3. Safe area insets on mobile shell render without clipping
- R4. Exchange order book with Buy/Sell tabs (success green / ruby red tokens), Sell tab as static "become a seller" prompt, empty-state handling
- R5. Listing table: advertiser name, price (MVR/USDT), available amount, trade limits, action button
- R6. Pre-filled test TRC-20 wallet in buy flow with "Demo wallet" disclaimer
- R7. Data-dense table on both viewports; mobile: horizontal scroll, sticky action column, name truncation, trade limits shorthand
- R8. Admin at hidden URL, no visible links
- R9. Passphrase gate with inline error on wrong entry, no lockout
- R10. Admin lists all sellers with name, slug, onboarding status
- R11. Reset clears `onboardingComplete`, `swipeClientId`, `swipeClientSecret`, deletes sessions; confirmation dialog before executing
- R12. Colored formatted Swipe API logs in dev server terminal

**Origin actors:** A1 (Presenter), A2 (Judge/audience), A3 (Seller), A4 (Buyer)
**Origin flows:** F1 (UI audit), F2 (Exchange browse/buy), F3 (Admin reset), F4 (Live Swipe demo)
**Origin acceptance examples:** AE1 (covers R1, R2), AE2 (covers R4, R5), AE3 (covers R6), AE4 (covers R9, R11), AE5 (covers R12)

---

## Scope Boundaries

- Landing page redesign — deferred, presenter handles separately
- Expo / React Native — dropped
- Real blockchain wallet integration
- Admin auth beyond passphrase
- Backend features beyond seller reset
- Performance optimization unrelated to visual fixes

---

## Context & Research

### Relevant Code and Patterns

- **Exchange browse**: `app/exchange/page.tsx` — card grid layout, uses `convex/exchange.ts:getActiveListings` (sorted by rate ascending). All data fields needed for order book rows already present.
- **Exchange buy flow**: `app/exchange/buy/[listingId]/page.tsx` — 4-step wizard (amount → wallet → review → payment). Step 2 collects buyer TRC-20 wallet with validation (`/^T[A-Za-z0-9]{33}$/`).
- **Convex stores**: `convex/stores.ts` — `getBySlug`, `verifyPin`, `completeOnboarding`, `setSwipeCredentials`. No `listAll` or `resetStore` mutations exist.
- **Convex sessions**: `convex/sessions.ts` — token-based, 24h TTL, `validate` returns `{storeId, storeSlug, storeName}`.
- **Swipe client**: `lib/swipe-client.ts` — centralized client with `createPayment`, `getPaymentStatus`, `simulatePayment`, `getWalletBalance`, `getTransactionHistory`, `getAccessToken`. All 10 API routes funnel through these functions. Handles `SWIPE_DEMO_MODE` internally.
- **Layout shells**: `components/shells/layout-router.tsx` — routes to MobileShell/DesktopShell (seller) or BuyerShell (buyer). No z-index conflicts in shells. BuyerShell mobile has no TopBar — exchange pages lack safe area handling.
- **Design tokens**: `app/globals.css` — amethyst (50-900), ruby (50-700), slate (50-900), semantic success (#059669), error (#dc2626). Font families, spacing, radius, shadow, motion tokens all defined.
- **Seller pages**: All use `p-4 md:p-6 max-w-Xxl mx-auto`. Dashboard has inconsistent grid cols (marketplace `grid-cols-4` vs crypto `grid-cols-3`). Orders table has 10 columns with `overflow-x-auto`.

### Institutional Learnings

- Dual-layout shell architecture: JS viewport detection at mount, no resize listeners. Mobile and desktop are separate DOM trees (`docs/solutions/architecture-patterns/dual-layout-shell-js-viewport-detection-2026-05-20.md`).
- Split responsibility: Convex owns data/auth, Next.js API routes own Swipe integration and SSE proxy (`docs/solutions/architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md`).

---

## Key Technical Decisions

- **Exchange is a restyle, not a rebuild**: The backend (`convex/exchange.ts`) already has all needed queries/mutations. `getActiveListings` returns listings sorted by rate — exactly what an order book needs. Only the browse page template changes from cards to a table.
- **Centralized logging in `lib/swipe-client.ts`**: All 10 Swipe API routes funnel through ~6 client functions. Instrumenting here catches every call with one change, vs adding logging to 10 separate route files.
- **Admin via Next.js API routes + new Convex mutations**: Admin passphrase checked at the API route level (not in Convex). New `stores.listAll` query and `stores.resetStore` mutation added to Convex. One schema change required: adding a `by_storeId` index to the sessions table for efficient session deletion during reset.
- **Admin URL: `/backstage`**: Less guessable than `/admin`, memorable for the presenter. No visible links anywhere.
- **Passphrase via `ADMIN_PASSPHRASE` env var**: Defaults to `"swipe2026"` in dev. Set in `.env.local`.

---

## Open Questions

### Resolved During Planning

- **Admin URL path**: `/backstage` — less guessable than `/admin`, easy to remember for demo
- **Passphrase mechanism**: Environment variable `ADMIN_PASSPHRASE` with dev default `"swipe2026"`
- **Test wallet address format**: TRC-20 (confirmed by existing validation in `stores.setCryptoWallet`)
- **Exchange component restructure**: Restyle only — card grid → table. Backend unchanged.

### Deferred to Implementation

- Exact test TRC-20 wallet address string to use (any valid-format address works)
- Per-page CSS fixes will be discovered during the visual audit pass

---

## Implementation Units

### U1. Swipe API demo logging

**Goal:** Add colored, formatted terminal logging to all Swipe API calls for live demo presentation.

**Requirements:** R12

**Dependencies:** None

**Files:**
- Modify: `lib/swipe-client.ts`

**Approach:**
- Create a `swipeLog(operation, details)` helper function at the top of `swipe-client.ts`
- Use ANSI color codes: green for success, yellow for pending, red for errors, cyan for info
- Format: `[SWIPE] HH:MM:SS ✓/✗/⏳ Operation — details`
- Instrument each exported function: log before the fetch (PENDING) and after (SUCCESS/ERROR)
- Replace existing bare `console.error` calls in API routes with the formatted logger
- Log operation type, amounts, merchant/order IDs, and status transitions

**Patterns to follow:**
- Existing `console.error` calls in `app/api/swipe/payments/create/route.ts` and `app/api/swipe/payments/simulate/route.ts` for error path locations

**Test scenarios:**
- Happy path: Payment creation logs `[SWIPE] ... ✓ Payment CREATED — MVR 150.00 — Order #abc123` with green coloring
- Happy path: Payment status check logs `[SWIPE] ... ⏳ Status CHECK — Payment #xyz — PENDING` with yellow coloring
- Happy path: Wallet balance query logs `[SWIPE] ... ✓ Wallet BALANCE — MVR 1,250.00` with green coloring
- Error path: Failed API call logs `[SWIPE] ... ✗ Payment CREATE FAILED — Connection refused` with red coloring
- Edge case: Demo mode calls log identically to real calls (logging is format-agnostic to mode)

**Verification:**
- Running the dev server and completing a checkout shows colored, timestamped payment lifecycle in the terminal
- All Swipe operations (create, status, simulate, wallet, history) produce formatted log lines

---

### U2. Admin backend — Convex mutations and API routes

**Goal:** Create the backend infrastructure for the admin dashboard: list all stores and reset a store's onboarding state.

**Requirements:** R8, R9, R10, R11

**Dependencies:** None

**Files:**
- Modify: `convex/stores.ts` — add `listAll` query and `resetStore` mutation
- Modify: `convex/schema.ts` — add `by_storeId` index to sessions table
- Modify: `convex/sessions.ts` — add `deleteByStoreId` mutation
- Create: `app/api/admin/stores/route.ts` — GET (list all) with passphrase check
- Create: `app/api/admin/stores/[storeId]/reset/route.ts` — POST (reset) with passphrase check
- Modify: `.env.local.example` — add `ADMIN_PASSPHRASE`

**Approach:**
- `stores.listAll`: New Convex query returning all stores with `name`, `slug`, `onboardingComplete`, `sellerType`, `swipeClientId` (presence only, not the value). No auth — protected at API route level.
- `stores.resetStore`: New Convex mutation taking `storeId`. Patches `onboardingComplete: undefined`, `swipeClientId: undefined`, `swipeClientSecret: undefined`. Then calls internal `sessions.deleteByStoreId`.
- `sessions.deleteByStoreId`: New internal mutation that queries sessions by `storeId` and deletes all matches.
- API routes check `Authorization` header against `process.env.ADMIN_PASSPHRASE || "swipe2026"`. Return 401 on mismatch. Use `convexServer` for Convex calls (existing pattern in other API routes).

**Patterns to follow:**
- `app/api/swipe/payments/create/route.ts` for API route structure with `convexServer`
- `convex/stores.ts:completeOnboarding` for mutation pattern with session auth (admin version skips session auth)
- `convex/sessions.ts:validate` for session table query pattern

**Test scenarios:**
- Happy path: GET `/api/admin/stores` with correct passphrase returns array of all stores with name, slug, onboarding status
- Happy path: POST `/api/admin/stores/[storeId]/reset` clears `onboardingComplete`, `swipeClientId`, `swipeClientSecret` and deletes all sessions for that store
- Error path: Any admin request without correct passphrase returns 401
- Error path: Reset with invalid storeId returns 404
- Integration: After reset, the seller's next login attempt (via `verifyPin`) succeeds but `onboardingComplete` is falsy, triggering the onboarding redirect in `app/seller/layout.tsx`
- Covers AE4: After passphrase + reset, seller's next login redirects to onboarding starting with Swipe API key submission

**Verification:**
- Calling the admin API with correct passphrase lists stores
- After calling reset, the target store's `onboardingComplete`, `swipeClientId`, `swipeClientSecret` are cleared and sessions deleted
- The seller hits the onboarding flow on next login

---

### U3. Exchange order book restyle — browse page

**Goal:** Transform the exchange browse page from a card grid to a Binance P2P-style order book with Buy/Sell tabs and data-dense table layout.

**Requirements:** R4, R5, R7

**Dependencies:** None

**Files:**
- Modify: `app/exchange/page.tsx` — full restyle from cards to order book table

**Approach:**
- Replace the card grid with a tabbed order book layout
- **Buy/Sell tabs**: Pill-style tab switcher at the top. Buy tab uses `--success` token (green), Sell tab uses ruby token (red). Buy selected by default.
- **Buy tab content**: Full-width table with columns: Advertiser (store name), Price (MVR/USDT in JetBrains Mono), Available (USDT amount), Limits (min–max shorthand), Action ("Buy USDT" button in success green)
- **Sell tab content**: Static prompt card — "Want to sell USDT?" with description text and a link to `/seller/exchange/new`
- **Empty state**: Table headers visible with centered "No listings available" message below
- **Mobile**: Table container with `overflow-x-auto`. Action column sticky with `position: sticky; right: 0`. Advertiser name truncated with `truncate` class. Trade limits as single `min-max` value.
- **Desktop**: Full table with comfortable column widths, no truncation needed
- **Header**: Add a page header with "P2P Trading" title and USDT/MVR market context (static or from listing data)
- Keep existing Convex query `getActiveListings` — already returns data sorted by rate ascending

**Patterns to follow:**
- Existing table pattern in `app/seller/orders/page.tsx` (desktop table with `overflow-x-auto`)
- Design system tokens from `app/globals.css` for colors, typography, spacing
- `components/shared/mvr-amount.tsx` for currency formatting

**Test scenarios:**
- Covers AE2: Buyer lands on exchange, sees Buy tab (green) selected with USDT listing table showing advertiser, price, available, limits, and "Buy USDT" button per row
- Happy path: Clicking Sell tab shows static "Want to sell USDT?" prompt with link to seller dashboard
- Happy path: Clicking "Buy USDT" navigates to `/exchange/buy/[listingId]`
- Edge case: Zero listings shows table headers with "No listings available" centered below
- Edge case: Mobile viewport shows horizontal scroll on table with sticky action column
- Edge case: Long advertiser name truncates with ellipsis on mobile

**Verification:**
- Exchange page renders as a data-dense order book table, not cards
- Buy/Sell tabs switch content correctly with appropriate color tokens
- Mobile horizontal scroll works with sticky action column
- Empty state renders cleanly

---

### U4. Exchange buy flow — wallet pre-fill and restyle

**Goal:** Pre-fill the buyer wallet address with a test TRC-20 address and add demo disclaimer. Restyle the buy flow to match the exchange's new visual language.

**Requirements:** R6

**Dependencies:** U3 (visual consistency with new exchange style)

**Files:**
- Modify: `app/exchange/buy/[listingId]/page.tsx`

**Approach:**
- In Step 2 (wallet entry), pre-fill the input with a test TRC-20 address (e.g., `TXYZ1234567890abcdefghijklmnopqrs` — 34 chars starting with T)
- Add a "Demo wallet — do not send real funds" label below the wallet input in a muted warning style
- Add a copy button next to the wallet input using `navigator.clipboard.writeText`
- Apply consistent typography: prices in JetBrains Mono, body text in DM Sans
- Match the amethyst/ruby/slate palette to the rest of the exchange section
- Keep the 4-step wizard flow intact — only visual changes + pre-fill

**Patterns to follow:**
- Existing wallet validation regex in the buy flow: `/^T[A-Za-z0-9]{33}$/`
- Design system spacing and color tokens

**Test scenarios:**
- Covers AE3: Buyer selects a listing, wallet field is pre-filled with test TRC-20 address, copy button present, "Demo wallet" label visible
- Happy path: Pre-filled address passes existing TRC-20 validation
- Happy path: Copy button copies the address to clipboard
- Happy path: User can clear and enter their own address
- Edge case: If user clears the field and enters an invalid address, validation error still shows

**Verification:**
- Buy flow Step 2 shows pre-filled wallet address with copy button and disclaimer
- The pre-filled address is a valid TRC-20 format (starts with T, 34 alphanumeric chars)
- The full purchase flow still works end-to-end with the pre-filled address

---

### U5. Admin dashboard UI

**Goal:** Build the passphrase-protected admin dashboard page for listing and resetting sellers during the demo.

**Requirements:** R8, R9, R10, R11

**Dependencies:** U2 (admin API routes)

**Files:**
- Create: `app/backstage/page.tsx`
- Create: `app/backstage/layout.tsx` — minimal layout, no shell (like landing page)

**Approach:**
- **Passphrase gate**: Full-screen centered form with passphrase input. On submit, store passphrase in component state and attempt to fetch stores via the admin API. On 401, show inline "Incorrect passphrase" error and clear the input. On success, show the admin panel. No session persistence — re-entry required if navigating away.
- **Seller list**: Table showing store name, slug, seller type, onboarding status (green badge for complete, gray for incomplete), Swipe credentials status (present/absent indicator). Use the existing `DashboardCard` visual style adapted for admin.
- **Reset action**: "Reset" button per seller row. On click, show confirmation dialog: "Are you sure? This will force re-onboarding for [Seller Name]." Confirm button in ruby/destructive style. On confirm, call reset API and refresh the list. Show success toast/inline feedback.
- **Layout**: No visible navigation links to `/backstage` anywhere. Layout bypasses the normal shell (similar to landing page routing in `layout-router.tsx`). Dark slate background with amethyst accents for a distinct "admin" feel.
- Store the passphrase in state and pass it as `Authorization` header on all admin API calls.

**Patterns to follow:**
- `app/page.tsx` (landing page) for shell-bypassing layout pattern
- `components/seller/dashboard-card.tsx` for card styling
- `app/seller/exchange/page.tsx` lines with the withdrawal confirmation modal for the confirmation dialog pattern

**Test scenarios:**
- Covers AE4: Presenter enters correct passphrase, sees seller list, clicks Reset, confirms, seller's next login goes to onboarding
- Happy path: Correct passphrase → seller list appears with all registered sellers
- Happy path: Reset button → confirmation dialog → confirm → seller state cleared → list refreshes with updated status
- Error path: Wrong passphrase → "Incorrect passphrase" inline error, input cleared
- Error path: Cancel on confirmation dialog → no action taken
- Edge case: No sellers registered → empty state message
- Edge case: Navigating away and back requires re-entering passphrase (no persistence)

**Verification:**
- `/backstage` shows passphrase form, not admin content
- Correct passphrase reveals seller list with accurate onboarding status
- Reset + confirm clears the seller's state (verifiable by checking the seller's next login redirects to onboarding)
- No link to `/backstage` exists anywhere in the app navigation

---

### U6. UI polish — seller pages

**Goal:** Audit and fix all seller dashboard pages for layout issues, design system consistency, and mobile/desktop shell integration.

**Requirements:** R1, R2, R3

**Dependencies:** None

**Files:**
- Modify: `app/seller/page.tsx` — dashboard
- Modify: `app/seller/products/page.tsx` — products list
- Modify: `app/seller/products/new/page.tsx` — product create/edit
- Modify: `app/seller/orders/page.tsx` — orders list
- Modify: `app/seller/payments/page.tsx` — payment links
- Modify: `app/seller/settings/page.tsx` — store settings
- Modify: `app/seller/social/page.tsx` — social publishing
- Modify: `app/seller/exchange/page.tsx` — seller exchange dashboard
- Modify: `app/seller/exchange/new/page.tsx` — create exchange listing
- Modify: `app/seller/onboarding/page.tsx` — onboarding flow
- Modify: `app/seller/login/page.tsx` — seller login

**Approach:**
- **Dashboard**: Fix inconsistent grid cols (marketplace `grid-cols-4` vs crypto `grid-cols-3` → unify to `grid-cols-2 md:grid-cols-4`). Verify stat cards use consistent spacing and typography.
- **All pages**: Verify `max-w-*` containment, `p-4 md:p-6` padding consistency, font families (Inter for headings, DM Sans for body, JetBrains Mono for prices/amounts), color palette adherence.
- **Mobile shell**: Verify content doesn't overlap TopBar (52px + safe area) or BottomTabBar (60px + safe area). Check `overflow-y-auto` on main content areas.
- **Tables (products, orders)**: Verify `overflow-x-auto` on table containers, check column widths at narrow desktop viewports.
- **Forms (product create, settings, onboarding)**: Verify input spacing, label typography, button styles match design system.
- Discovered issues will be fixed inline during the audit — the exact CSS changes are implementation-time discoveries.

**Patterns to follow:**
- Existing seller page patterns for consistency
- Design tokens in `app/globals.css`

**Test scenarios:**
- Covers AE1: Seller dashboard mobile products page renders without overlapping bottom tab bar or top bar
- Happy path: Each seller page renders cleanly on both mobile (375px) and desktop (1280px) viewports
- Happy path: All headings use Inter, body text uses DM Sans, prices use JetBrains Mono
- Edge case: Orders table with 10 columns scrolls horizontally on narrow viewports without content clipping
- Edge case: Product cards on mobile don't overlap with bottom tab bar

**Verification:**
- Visual spot-check of every seller page on both mobile and desktop — no overlapping elements, clipped content, or design system violations

---

### U7. UI polish — buyer pages and shell fixes

**Goal:** Audit and fix all buyer-facing pages, plus fix the BuyerShell mobile safe area gap discovered during research.

**Requirements:** R1, R2, R3

**Dependencies:** U3, U4 (exchange pages already restyled)

**Files:**
- Modify: `components/shells/buyer-shell.tsx` or `components/mobile/buyer-bottom-bar.tsx` — add safe area handling
- Modify: `app/marketplace/page.tsx` — marketplace browse
- Modify: `app/shop/[storeSlug]/page.tsx` — store page
- Modify: `app/shop/[storeSlug]/product/[productId]/page.tsx` — product detail page
- Modify: `app/checkout/[orderId]/page.tsx` — checkout
- Modify: `app/success/[orderId]/page.tsx` — success page

**Approach:**
- **BuyerShell mobile safe area**: The BuyerShell mobile layout has no TopBar, so exchange and marketplace pages render edge-to-edge at the top. Add `padding-top: env(safe-area-inset-top)` to the BuyerShell's main content area on mobile, or add a minimal top spacer.
- **BuyerBottomBar**: Verify it has the same safe area inset handling as the seller BottomTabBar (`paddingBottom: env(safe-area-inset-bottom)`).
- **Marketplace**: Check product grid card spacing, image aspect ratios, search/filter UI consistency.
- **Store/product pages**: Check variant selectors, quantity inputs, delivery form layout.
- **Checkout/success**: Check payment QR display, status polling UI, order confirmation layout.
- Apply same design system audit as seller pages: typography, colors, spacing, radius.

**Patterns to follow:**
- `components/mobile/top-bar.tsx` and `components/mobile/bottom-tab-bar.tsx` for safe area handling patterns
- Seller page audit approach from U6

**Test scenarios:**
- Happy path: Marketplace page renders product grid cleanly on both viewports
- Happy path: Checkout page shows payment QR/link without overlap
- Edge case: BuyerShell mobile content doesn't clip behind device notch/status bar
- Edge case: BuyerBottomBar has safe area bottom padding on devices with home indicators
- Edge case: Success page renders cleanly after payment completion

**Verification:**
- Visual spot-check of every buyer page on both mobile and desktop
- BuyerShell mobile content has safe area top padding — no content behind the notch
- No overlapping elements or design system violations on buyer pages

---

## System-Wide Impact

- **Interaction graph**: Admin reset touches Convex stores + sessions tables. The seller's existing logged-in tabs will lose their session on next API call after reset. The `seller/layout.tsx` onboarding check will catch the cleared state on next page load.
- **Error propagation**: Swipe logging wraps existing error handling — it should not swallow errors. Admin API errors return appropriate HTTP status codes (401, 404, 500).
- **State lifecycle risks**: Admin reset during an active seller session means the seller's in-memory state (localStorage `storeSlug`, `storeName`) may persist while the backend session is deleted. This is acceptable for a demo — the seller will be redirected to login on the next authenticated API call.
- **API surface parity**: No existing APIs change. Two new admin API routes are added. Two new Convex mutations are added. All are additive.
- **Unchanged invariants**: The existing seller login flow, buyer purchase flow, Swipe payment integration, and Convex real-time subscriptions are not modified. Exchange backend queries and mutations remain unchanged.

---

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| UI polish is open-ended and could consume excessive time | Page inventory in R1 bounds the scope; fix obvious issues, don't chase perfection |
| Exchange table may look cramped on small mobile screens | Horizontal scroll + sticky action column + truncation provide fallbacks |
| Admin passphrase visible in env vars or terminal | Acceptable for hackathon — not a production deployment |
| Swipe logging adds overhead to API calls | Logging is synchronous console output — negligible overhead |

---

## Sources & References

- **Origin document:** `docs/brainstorms/2026-05-21-ui-polish-exchange-admin-requirements.md`
- Design system: `app/globals.css`
- Exchange backend: `convex/exchange.ts`
- Store schema: `convex/schema.ts`
- Swipe client: `lib/swipe-client.ts`
- Layout architecture: `docs/solutions/architecture-patterns/dual-layout-shell-js-viewport-detection-2026-05-20.md`
