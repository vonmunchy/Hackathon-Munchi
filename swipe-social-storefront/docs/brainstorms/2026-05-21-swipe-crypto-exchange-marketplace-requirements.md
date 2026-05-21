---
date: 2026-05-21
topic: crypto-exchange-and-marketplace-upgrade
---

# Swipe Platform: P2P Crypto Exchange + Unified Marketplace

## Summary

Two-pillar platform combining a unified marketplace (aggregating all seller storefronts with search, categories, and seller filtering) and a P2P USDT crypto exchange — both powered by Swipe as the instant MVR payment rail. Sellers onboard into one or both pillars. Buyers browse freely with zero registration, providing details only at checkout. All crypto operations simulated for hackathon.

---

## Problem Frame

In the Maldives, crypto trading happens through informal channels — Telegram groups, WhatsApp DMs, and word-of-mouth. Buyers send bank transfers to strangers and hope the USDT arrives. Sellers post wallet addresses publicly and trust that transfer slips are real. Scams are common, verification is manual, and there is no escrow or instant settlement.

Meanwhile, social sellers on Instagram and Facebook still manage orders via DMs with no unified discovery surface — buyers must find individual sellers through social media rather than browsing a marketplace.

Swipe's instant payment confirmation solves both problems: for marketplace purchases, no more fake transfer slips; for crypto exchanges, Swipe payment confirmation triggers automatic USDT release from escrow — no trust required between buyer and seller.

---

## Actors

- A1. Marketplace Seller: Onboards to sell physical products via a storefront, publishes to the unified marketplace
- A2. Crypto Seller: Onboards to sell USDT at their chosen rate, deposits into platform escrow
- A3. Buyer: Browses marketplace and/or crypto exchange without registration, purchases at checkout
- A4. Platform (Swipe): Manages escrow wallets, confirms payments, releases USDT automatically

---

## Key Flows

- F1. Crypto Seller Lists USDT
  - **Trigger:** Seller navigates to "Create Listing" in their crypto dashboard
  - **Actors:** A2, A4
  - **Steps:**
    1. Seller specifies amount of USDT to sell and their rate (MVR per USDT)
    2. Seller chooses whether to allow partial purchases (default: yes)
    3. Platform generates a fresh wallet address for this listing
    4. Seller "deposits" USDT to the generated address (simulated — instant confirmation)
    5. Listing goes live with a publicly viewable wallet/escrow balance
  - **Outcome:** USDT is escrowed and the listing is visible to buyers
  - **Covered by:** R5, R6, R7, R8

- F2. Buyer Purchases USDT
  - **Trigger:** Buyer selects a listing and clicks "Buy"
  - **Actors:** A3, A4
  - **Steps:**
    1. Buyer enters desired amount (if partial allowed) or confirms full amount
    2. Buyer provides their USDT wallet address (TRC20/ERC20 format)
    3. Platform generates Swipe payment link for the MVR equivalent
    4. Buyer is redirected to Swipe payment page (5-second auto-return or manual)
    5. Platform detects payment confirmation via polling/SSE
    6. Platform transfers USDT from escrow wallet to buyer's provided address (simulated)
    7. Buyer sees confirmation with transaction details
  - **Outcome:** Buyer has USDT in their wallet; seller's MVR is credited; escrow balance reduced
  - **Covered by:** R9, R10, R11, R12, R13

- F3. Seller Withdraws Unsold USDT
  - **Trigger:** Seller clicks "Withdraw" on an active listing
  - **Actors:** A2, A4
  - **Steps:**
    1. Seller requests withdrawal of remaining escrowed USDT
    2. Platform transfers USDT back to seller's original wallet (simulated)
    3. Listing is deactivated/removed
  - **Outcome:** Seller's USDT returned; listing no longer visible to buyers
  - **Covered by:** R14

- F4. Buyer Browses Unified Marketplace
  - **Trigger:** Buyer navigates to marketplace
  - **Actors:** A3
  - **Steps:**
    1. Buyer sees aggregated products from all marketplace sellers
    2. Buyer can search by text, filter by category, or browse by seller
    3. Buyer selects a product and proceeds to purchase (existing checkout flow)
  - **Outcome:** Buyer discovers and purchases products across all stores from one surface
  - **Covered by:** R15, R16, R17, R18

---

## Requirements

**Onboarding**
- R1. After Meta OAuth login, seller chooses their path: "Marketplace", "Crypto Exchange", or "Both"
- R2. Marketplace path follows existing onboarding (phone + Swipe API credentials + test products)
- R3. Crypto Exchange path collects seller's USDT wallet address (TRC20 format, validated at entry). This is the address they deposit from and withdraw to
- R4. Sellers who choose "Both" complete both flows and see both dashboard sections

**Crypto Exchange — Seller**
- R5. Seller can create a listing specifying: USDT amount, rate (MVR per USDT), and partial-buy toggle (default: allowed)
- R6. Platform generates a unique wallet address per listing for escrow
- R7. Seller "deposits" USDT to the escrow address (simulated instant confirmation for hackathon)
- R8. Each listing displays its escrow wallet address and balance publicly (transparency)
- R14. An authenticated seller can withdraw remaining escrowed USDT only from listings they own, back to their original wallet. Withdrawal is blocked if any purchase for that listing has a confirmed Swipe payment pending USDT release

**Crypto Exchange — Buyer**
- R9. Buyers browse all active listings showing: seller name, USDT amount available, rate, total MVR cost
- R10. Buyer enters desired USDT amount (constrained by listing's available balance and partial-buy setting)
- R11. Buyer provides their USDT wallet address at purchase time (no account required). Address must pass TRC20 format validation (starts with T, 34 characters) before payment link is generated
- R12. Platform generates a Swipe payment link for the calculated MVR amount
- R13. Upon Swipe payment confirmation (verified server-side against Swipe API), platform records the USDT transfer (simulated transaction hash), reduces the listing's available balance, and shows the buyer a confirmation screen. Balance check and deduction happen in a single Convex mutation to prevent double-spend

**Unified Marketplace**
- R15. A new marketplace route aggregates products from all active stores
- R16. Buyers can search products by text (name, description)
- R17. Buyers can filter products by category
- R18. Buyers can browse/filter by seller (store)
**Landing Page**
- R20. Landing page communicates both value props: marketplace commerce and crypto exchange
- R21. Two clear entry paths for sellers ("Sell Products" and "Sell Crypto") and two for buyers ("Shop" and "Buy Crypto")
- R22. Theme centers on "what wasn't possible before is now possible with Swipe" — instant payments eliminating trust issues

**Layout & Navigation**
- R23. All new surfaces (marketplace, crypto exchange) have distinct mobile (app-like) and desktop layouts consistent with existing shell pattern
- R24. Mobile navigation includes access to both marketplace and crypto exchange sections
- R25. Desktop navigation includes sidebar entries for both sections

---

## Acceptance Examples

- AE1. **Covers R5, R7.** Given a crypto seller is onboarded, when they create a listing for 100 USDT at 25.50 MVR/USDT with partial buys allowed, the platform generates a wallet address, seller confirms deposit, and the listing appears with "100 USDT available" and the escrow address visible.

- AE2. **Covers R10, R12, R13.** Given a listing has 100 USDT available with partial buys allowed, when a buyer requests 50 USDT and provides their wallet, the platform generates a Swipe payment for 1,275 MVR (50 x 25.50). After payment confirms, the listing shows "50 USDT available" and buyer sees a confirmation with simulated transaction hash.

- AE3. **Covers R5, R10.** Given a listing has partial buys disabled, when a buyer attempts to purchase, they can only buy the full listed amount — no amount input field is shown.

- AE4. **Covers R14.** Given a listing has 50 USDT remaining after partial sales, when the seller withdraws, the 50 USDT returns to their wallet and the listing disappears from the exchange.

- AE5. **Covers R1, R4.** Given a new seller completes Meta OAuth, when they select "Both" on the path selection screen, they complete marketplace onboarding (phone, Swipe creds, test products) AND crypto onboarding (wallet address), then see both marketplace and crypto sections in their dashboard.

- AE6. **Covers R15, R16, R18.** Given multiple stores have active products, when a buyer searches "abaya" on the marketplace, results show matching products across all stores with the seller name visible on each card.

---

## Success Criteria

- A hackathon judge can watch a seller list USDT, then a buyer purchase it via Swipe, and understand immediately that the platform eliminates trust-based scams
- The marketplace feels like a real marketplace (search, categories, multiple sellers) — not just a single-store shop
- The landing page clearly communicates both pillars without feeling cluttered
- Mobile experience feels app-native; desktop experience uses screen space effectively
- The end-to-end crypto flow takes under 60 seconds from buyer selecting a listing to receiving "USDT transferred" confirmation

---

## Scope Boundaries

### Deferred for later

- KYC/AML verification for crypto sellers
- Dispute resolution between buyers and sellers
- Buyer accounts and purchase history
- Real-time market rate feeds
- Multi-currency support (only USDT for now)
- Seller reputation/rating system
- Admin panel for exchange management
- Real blockchain integration (post-hackathon enhancement — all crypto operations are explicitly simulated for the demo)

### Constraints (not deliverables)

- Existing per-store pages (`/shop/[slug]`) continue to work alongside the marketplace (additive, no breaking changes)

### Outside this product's identity

- Centralized exchange (order book, market makers) — this is P2P only
- Custodial wallet service — platform holds escrow temporarily, not long-term
- DeFi features (staking, lending, yield)
- Fiat-to-fiat remittance

---

## Key Decisions

- **Partial buys default-on, seller-controlled**: Sellers can toggle whether their listing accepts partial purchases. Default is allowed. This gives flexibility without adding buyer confusion.
- **No buyer registration**: Zero friction for buyers. Wallet address collected at purchase time only. Trades hackathon demo speed for repeat-buyer convenience.
- **Simulated crypto operations**: All wallet generation, deposits, transfers, and balances are simulated in Convex. Wallet addresses are generated strings that look real (TRC20 format). This demonstrates the concept without blockchain complexity.
- **Unified marketplace as addition, not replacement**: Existing per-store pages remain. The new `/marketplace` route is an aggregation layer on top of existing store/product data.
- **Swipe payment redirect with auto-return**: Payment opens Swipe web page with a 5-second countdown timer for auto-redirect back. Manual "Return to app" button as fallback for mobile browsers that block auto-redirects.
- **Platform-level Swipe credentials for crypto exchange**: Crypto exchange payments use platform-level Swipe credentials (environment fallback), not per-seller credentials. MVR flows buyer → platform → seller is implicit (for hackathon, MVR credit to seller is simulated alongside USDT transfer).
- **Reservation model for partial purchases**: When a buyer initiates a purchase, the requested USDT amount is soft-reserved for 5 minutes. If payment does not confirm within that window, the reservation expires and balance is restored. R10 constrains against unreserved available balance.
- **Buyer-facing routes are public**: `/marketplace` and `/exchange` are public buyer-facing routes using the buyer shell (no auth required). Seller-facing crypto management lives under `/seller/exchange` using the existing seller shell.
- **Navigation model**: Buyer-facing surfaces (marketplace, exchange) use a separate buyer shell with their own navigation. The seller bottom-tab/sidebar gains an "Exchange" entry for crypto management. Buyers see: Home, Marketplace, Exchange, (no auth-required tabs).
- **Success criteria priority**: P0 = crypto exchange end-to-end demo (novel differentiator). P1 = marketplace aggregation with search. P2 = polished landing page. This gives a clear cut line under time pressure.

---

## Dependencies / Assumptions

- Existing Swipe payment integration (create payment, poll status) works and can be reused for crypto purchases
- Convex schema can be extended with new tables for exchange listings, wallets, and transactions
- The existing session/auth system supports the new onboarding paths without breaking changes
- Demo mode (SWIPE_DEMO_MODE) handles all simulated payments for the exchange as well

---

## Outstanding Questions

### Deferred to Planning

- [Affects R6][Technical] TRC20 address generation format for escrow wallets (confirmed TRC20 only — starts with T, 34 chars)
- [Affects R12][Technical] Whether to reuse the existing QR-based Swipe checkout or create a payment-link-only flow for crypto purchases
- [Affects R15, R16][Technical] How to efficiently query products across all stores — Convex searchIndex vs client-side filtering for hackathon scale
- [Affects R1][Technical] Schema addition needed: `sellerType` field on stores table to persist path selection (marketplace/crypto/both)
- [Affects R5][Technical] Reservation expiry mechanism — Convex scheduled function or client-side timer with server validation
- [Affects F2][Technical] Polling interval and timeout for payment confirmation detection (must support 60-second success criterion)
