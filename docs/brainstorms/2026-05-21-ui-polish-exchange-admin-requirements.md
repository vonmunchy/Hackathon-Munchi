---
date: 2026-05-21
topic: ui-polish-exchange-admin
---

# UI Polish, Binance P2P Exchange, and Admin Dashboard

## Summary

Full UI quality-of-life pass across all seller and buyer pages to fix layout issues and consistently apply the amethyst/ruby/slate design system, a Binance P2P-style exchange redesign with full order book layout, a passphrase-protected hidden admin dashboard for demo control, and verbose Swipe API logging in the dev server terminal for live presentation.

---

## Problem Frame

The web app is heading into hackathon judging. The current UI has minor but visible issues — overlapping elements, inconsistent spacing, layout jank — that undermine the polish judges expect. The exchange section works functionally but doesn't evoke the "real crypto platform" feel that a Binance P2P-style layout would. During the demo, the presenter needs to reset sellers on the fly (forcing re-onboarding to show the Swipe API key flow) and show Swipe payment simulation in the terminal to prove the system works end-to-end. There's no admin interface for any of this today.

---

## Actors

- A1. Presenter (hackathon demo): Navigates the app live, resets sellers via admin, shows terminal output
- A2. Judge / audience: Evaluates visual quality, exchange UX, and live payment simulation
- A3. Seller (demo persona): Goes through onboarding, manages products, receives payments
- A4. Buyer (demo persona): Browses marketplace, uses exchange, completes purchases

---

## Key Flows

- F1. Full UI audit and fix
  - **Trigger:** Developer initiates polish pass
  - **Actors:** A2, A3, A4 (affected)
  - **Steps:** Audit every seller and buyer page for overlapping elements, inconsistent spacing, misaligned components, and design system violations. Fix each issue. Verify both mobile and desktop shells.
  - **Outcome:** All pages render cleanly with consistent amethyst/ruby/slate design tokens, no visible layout issues
  - **Covered by:** R1, R2, R3

- F2. Exchange browse and buy
  - **Trigger:** Buyer navigates to exchange
  - **Actors:** A4
  - **Steps:** Buyer sees order book with buy/sell tabs, green/red color coding, listing table with advertiser info and trade limits. Selects a listing, enters amount, copies test wallet address, initiates purchase.
  - **Outcome:** Exchange feels like Binance P2P — data-dense, professional, trustworthy
  - **Covered by:** R4, R5, R6, R7

- F3. Admin resets a seller
  - **Trigger:** Presenter navigates to hidden admin URL and enters passphrase
  - **Actors:** A1
  - **Steps:** Presenter goes to admin URL, enters passphrase, sees list of sellers, selects one, clicks reset/remove. Seller's onboarding state is cleared. Next time that seller logs in, they go through full onboarding including Swipe API key submission.
  - **Outcome:** Seller is forced back to onboarding; product and order data preserved
  - **Covered by:** R8, R9, R10, R11

- F4. Live Swipe payment demo
  - **Trigger:** A buyer completes a purchase while presenter has dev server terminal visible
  - **Actors:** A1, A2
  - **Steps:** Payment is created, terminal shows colored log output with timestamps, amounts, merchant IDs, and status transitions (created → processing → completed).
  - **Outcome:** Judges see real-time proof the payment system works
  - **Covered by:** R12

---

## Requirements

**UI polish**
- R1. Fix all overlapping elements, z-index conflicts, and layout overflow issues across the following pages in both mobile and desktop shells:
  - Buyer: `/marketplace`, `/exchange`, `/exchange/buy/[listingId]`, `/shop/[storeSlug]`, `/shop/[storeSlug]/product/[productId]`, `/checkout/[orderId]`, `/success/[orderId]`
  - Seller: `/seller`, `/seller/products`, `/seller/products/new`, `/seller/orders`, `/seller/payments`, `/seller/settings`, `/seller/social`, `/seller/exchange`, `/seller/exchange/new`, `/seller/onboarding`, `/seller/login`
- R2. Ensure consistent application of the amethyst/ruby/slate design system — typography (Inter/DM Sans/JetBrains Mono), spacing tokens, border radius, shadows, and color palette
- R3. Verify safe area insets on mobile shell (top bar, bottom tabs) render without content clipping or overlap

**Exchange — Binance P2P style**
- R4. Exchange landing page displays a full order book layout with Buy/Sell tab switcher using the design system's semantic success (green) and ruby (red) tokens. Buy tab shows active USDT listings. Sell tab displays a static prompt directing users to the seller dashboard to create a listing. When a tab has zero listings, show table headers with a centered empty-state message ("No listings available" or equivalent)
- R5. Listing table shows columns: advertiser name, price (MVR/USDT), available amount, trade limits (min–max), and a prominent action button
- R6. The buy flow pre-fills a test TRC-20 wallet address (e.g., `TXqH...abc123`) as the buyer's destination wallet, with a copy button for convenience. The address is labeled "Demo wallet — do not send real funds"
- R7. Both mobile and desktop render the order book table layout (data-dense on both viewports). On mobile, the table scrolls horizontally with a sticky action column. Advertiser name truncates with ellipsis. Trade limits collapse to a single "min–max" shorthand

**Admin dashboard**
- R8. Admin dashboard accessible at a hidden URL with no visible links anywhere in the app
- R9. Access requires entering a hardcoded passphrase before any admin content is shown. On wrong passphrase, show an inline "Incorrect passphrase" error and clear the input field. No rate limiting or lockout — the presenter may need multiple quick attempts during a live demo
- R10. Admin view lists all registered sellers with their store name, slug, and onboarding status
- R11. "Reset seller" action clears `onboardingComplete`, `swipeClientId`, and `swipeClientSecret` on the store record, and deletes all active sessions for that store — forcing full re-onboarding (including Swipe API key submission) on next login. Product, order, and store configuration data are preserved. Reset requires a confirmation step ("Are you sure? This will force re-onboarding for [Seller Name]") before executing

**Swipe demo logging**
- R12. When Swipe API calls are made (payment creation, status checks, wallet queries), the Next.js dev server terminal outputs colored, formatted logs showing: operation type, timestamp, amounts, merchant/order IDs, and status transitions

---

## Acceptance Examples

- AE1. **Covers R1, R2.** Given the seller dashboard on mobile, when viewing the products page, all product cards render within the viewport without overlapping the bottom tab bar or top bar, using the correct font families and spacing tokens.
- AE2. **Covers R4, R5.** Given the exchange page, when a buyer lands on it, they see a Buy tab (green) selected by default with a table of USDT listings showing advertiser, price, available amount, limits, and a "Buy USDT" button per row.
- AE3. **Covers R6.** Given the exchange buy flow, when a buyer selects a listing, the buyer's destination wallet field is pre-filled with a test TRC-20 address (e.g., `TXqH...abc123`) with a copy button and a "Demo wallet — do not send real funds" label.
- AE4. **Covers R9, R11.** Given the admin dashboard, when the presenter enters the correct passphrase and clicks "Reset" on a seller, that seller's next login redirects them to the onboarding flow starting with Swipe API key submission.
- AE5. **Covers R12.** Given a buyer completing checkout, when the Swipe payment API is called, the dev server terminal shows a colored log line like `[SWIPE] 14:32:05 ✓ Payment CREATED — MVR 150.00 — Order #abc123`.

---

## Success Criteria

- Every seller and buyer page in both mobile and desktop shells passes a visual spot-check with no overlapping elements, clipped content, or design system violations
- A judge browsing the exchange section gets a "this looks like Binance" first impression within 3 seconds
- The presenter can reset a seller and demonstrate the full re-onboarding flow in under 30 seconds during the demo
- Terminal output during a live purchase clearly shows the payment lifecycle to a non-technical audience

---

## Scope Boundaries

- Landing page redesign — deferred, presenter will handle separately
- Expo / React Native mobile app — dropped, web-only going forward
- Real blockchain wallet integration or on-chain transactions
- Admin authentication beyond a simple hardcoded passphrase
- New backend features beyond seller reset (no order management, no analytics in admin)
- Performance optimization or code refactoring unrelated to visual fixes

---

## Key Decisions

- **Full order book table on both viewports (not cards on mobile):** Maximizes the "real crypto platform" impression for judges, even though a card layout would be more conventionally mobile-friendly
- **Passphrase-only admin auth:** Hackathon demo context — knowing the URL + passphrase is sufficient security. No user accounts, no sessions for admin.
- **Seller reset preserves data:** Only onboarding/session state is cleared. Products, orders, and store config remain intact so the demo can show continuity.

---

## Dependencies / Assumptions

- The existing amethyst/ruby/slate design system in `globals.css` is the source of truth for all visual tokens
- Convex backend has seller records with an onboarding status field that can be reset
- A new Convex query is needed for admin that returns all stores without per-store session authentication — this query should be protected at the API route level by the admin passphrase check, not by session-based auth
- Swipe API calls already flow through Next.js API routes where logging can be added
- Demo mode (`SWIPE_DEMO_MODE=true`) is the expected runtime for the presentation

---

## Outstanding Questions

### Deferred to Planning

- [Affects R8][Technical] What should the admin URL path be? (e.g., `/admin`, `/backstage`, something less guessable)
- [Affects R9][Technical] What passphrase to use? Can be set via environment variable for flexibility.

- [Affects R4][Needs research] Current exchange component structure — how much can be restyled vs needs full rebuild for order book layout
