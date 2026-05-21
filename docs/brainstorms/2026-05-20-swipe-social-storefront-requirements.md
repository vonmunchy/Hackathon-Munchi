---
date: 2026-05-20
topic: swipe-social-storefront
---

# Swipe Social Storefront — Hackathon MVP Requirements

## Summary

A mobile-first social commerce platform for Maldives-based Instagram and Facebook sellers, powered by the BML Swipe payment API. Sellers manage products, inventory, and orders through a dashboard; buyers browse a storefront shared via social media, pay with Swipe (QR on desktop, payment link on mobile), and provide delivery details for Male'/Hulhumale'. The platform replaces manual transfer-slip verification with automatic payment confirmation and real-time inventory management.

---

## Problem Frame

Many small businesses in Maldives sell through Instagram and Facebook. The current workflow is entirely manual: sellers post product images, buyers DM, sellers check stock by memory, buyers send bank transfer slips, and sellers manually verify payment by checking their bank app. This creates fake-slip fraud risk, delayed order confirmation, overselling from poor stock tracking, no payment-to-order reconciliation, and no structured order records. Sellers spend significant time on operations that could be automated, and buyers experience friction and uncertainty throughout the purchase process.

---

## Actors

- A1. **Buyer**: A Maldives-based consumer who discovers products on Instagram/Facebook, visits the storefront, selects products, pays via Swipe, and provides delivery details for Male'/Hulhumale'.
- A2. **Seller**: A small business owner who manages products, inventory, social captions, and orders through the seller dashboard. Currently sells via Instagram/Facebook DMs.
- A3. **Swipe API**: BML's payment platform (mock API for hackathon). Handles payment creation, QR code generation, payment link generation, SSE status streaming, webhook delivery, and transaction history.
- A4. **Hackathon judges**: Evaluate the demo for innovation, Swipe API depth, and production readiness.

---

## Key Flows

- F1. **Buyer purchase flow**
  - **Trigger:** Buyer clicks a storefront link shared on Instagram/Facebook
  - **Actors:** A1, A3
  - **Steps:**
    1. Buyer lands on the storefront, browses products (with category filtering)
    2. Buyer selects a product, chooses variant and quantity
    3. Buyer clicks "Buy Now" — app checks stock availability
    4. Buyer fills delivery form: name, phone, location (Male'/Hulhumale'), address, delivery time preference
    5. App creates a Swipe payment (QR type on desktop, LINK type on mobile) with product name + order ID as description
    6. Checkout page displays QR code or payment link, opens SSE stream for real-time status
    7. Buyer completes payment (in production via Swipe app; in demo via mock's `/pay/{reference}/complete`)
    8. SSE stream fires COMPLETED; app confirms order and deducts stock
    9. Success page shows order confirmation, amount paid, order ID, and "Message Seller" buttons (WhatsApp/Instagram DM deep links)
  - **Outcome:** Order is paid, stock is reduced, seller sees the order on their dashboard
  - **Covered by:** R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12

- F2. **Seller product and social workflow**
  - **Trigger:** Seller opens the dashboard to add/manage products
  - **Actors:** A2
  - **Steps:**
    1. Seller adds a product with name, price (MVR), description, category, and up to 5 images (or placeholder URLs for MVP)
    2. Seller adds variants (size/color/option) with stock quantities
    3. Seller navigates to social caption generator, selects a product
    4. App generates Instagram caption, Facebook caption, and hashtags (template-based, with optional AI)
    5. Seller copies captions and shares on their social channels with the storefront/product link
  - **Outcome:** Product is live on the storefront with social-ready captions
  - **Covered by:** R13, R14, R15, R16, R17

- F3. **Seller order management flow**
  - **Trigger:** A new paid order appears on the dashboard
  - **Actors:** A2
  - **Steps:**
    1. Seller sees "new orders" indicator on the dashboard
    2. Seller views order details: product, variant, quantity, amount, buyer contact, delivery address, delivery time, Swipe reference
    3. Seller marks order as "shipped" or "delivered"
    4. Seller views Swipe wallet balance and transaction history with fee breakdowns
  - **Outcome:** Seller has full visibility into orders, payments, and fulfillment status
  - **Covered by:** R18, R19, R20, R21, R22, R23

- F4. **Hackathon demo flow**
  - **Trigger:** Demo presentation to judges
  - **Actors:** A2, A4
  - **Steps:**
    1. Show "before vs after" comparison: old manual flow vs Swipe-powered flow
    2. Show seller dashboard with demo store "Island Finds MV"
    3. Show product management with variants and stock
    4. Generate social captions for a product
    5. Open buyer storefront on mobile viewport
    6. Complete a purchase: select product, fill delivery, pay via Swipe mock
    7. Show real-time payment confirmation via SSE
    8. Show inventory auto-deduction on seller dashboard
    9. Show Swipe wallet balance and transaction history
    10. Show impact metrics (time saved, fraud eliminated, reconciliation automated)
  - **Outcome:** Judges see a complete, polished, end-to-end social commerce flow powered by Swipe
  - **Covered by:** R24, R25, R26, R27

---

## Requirements

**Buyer storefront**

- R1. Storefront page displays products in a grid with image, name, price (MVR), and stock status, with category-based filtering or tabs.
- R2. Product detail page shows product images, name, price, description, variant selector, quantity selector, stock availability per variant, and a "Buy Now" button. Button is disabled when the selected variant is out of stock.
- R3. When a buyer shares a storefront or product link on social media, the page renders Open Graph meta tags (title, description, image, price) so the link previews correctly on Instagram, Facebook, WhatsApp, and other platforms.
- R4. Buyers can share product links via share buttons (WhatsApp, copy link) on the product page.

**Delivery and checkout**

- R5. After clicking "Buy Now" and passing stock validation, the buyer fills a delivery form: name, phone number, delivery location (Male' or Hulhumale' selector), delivery address (free text), and delivery time preference (free text, e.g., "evening", "after 5pm").
- R6. Always create a QR-type Swipe payment (guaranteed mock support). The checkout page displays **both** the QR code image (base64 PNG from `qr_data`) **and** a "Pay with Swipe" button linking to the mock's pay page at `/pay/{shortCode}` on all viewports (desktop and mobile). On desktop, the QR is prominently displayed with the link as a secondary option. On mobile, the link button is prominently displayed with a smaller QR shown below for flexibility. This ensures buyers always have both payment options regardless of device.
- R7. The checkout page receives real-time payment status updates without page refresh. All Swipe API communication is server-to-server — the Next.js backend proxies the Swipe SSE stream via a server-side API route (e.g., `/api/checkout/[orderId]/stream`) that authenticates to Swipe with the cached merchant OAuth token and re-emits events to the browser client. The buyer's browser never connects directly to the Swipe API.
- R8. When the SSE stream reports COMPLETED status, the app confirms the order (sets status to paid) and deducts stock (`stock_available -= quantity`, `stock_sold += quantity`) in a single operation.
- R9. If the payment expires or is cancelled (reported via SSE), the checkout page shows an appropriate message and offers to retry.
- R10. A "Simulate Payment" fallback button is available on the checkout page that fires a server-side POST to the mock's `/pay/{shortCode}/complete` endpoint (where shortCode is the payment's `short_code`/`reference` value). The endpoint returns a 303 redirect, not JSON — the button should fire-and-forget. For demo purposes and as a fallback if SSE fails.

**Post-payment buyer experience**

- R11. Success page shows: payment confirmed message, order ID, product summary, amount paid in MVR, and delivery details summary.
- R12. Success page includes "Message Seller" buttons that deep-link to: (a) WhatsApp chat with the seller's phone number pre-filled with order reference, and (b) Instagram DM to the seller's handle. Also includes a "Back to Store" button. (Depends on R16: reads seller WhatsApp number and Instagram handle from the store profile. If not configured, deep links fall back to a placeholder or are omitted.)

**Seller product management**

- R13. Sellers can add, edit, and archive products. Each product has: name, description, base price (MVR), category, status (draft/active/archived), and up to 5 image URLs.
- R14. Each product has one or more variants with: variant name (e.g., size, color), price override (optional), and stock quantity (`stock_available`, `stock_sold`).
- R15. Product list view shows: image, name, price, variant count, total available stock, total sold, and status.
- R16. Seller can update store profile: store name, description, Instagram username, Facebook page URL, WhatsApp number, and logo URL.

**Social caption generator**

- R17. Seller selects a product and generates Instagram caption, Facebook caption, and hashtag suggestions. Captions include the product name, price, variant info, stock urgency (if low), and a storefront/product link. Template-based by default; AI-powered if an API key is configured. Copy button for each caption.

**Seller order and fulfillment management**

- R18. Order list shows: order ID, buyer name, product/variant, quantity, amount (MVR), payment status, order status, Swipe transaction reference, delivery location, and created timestamp (in Maldives timezone, UTC+5).
- R19. When a new paid order arrives, the seller dashboard shows a visual indicator (badge/count on the orders section).
- R20. Sellers can update order status through a fulfillment workflow: paid -> shipped -> delivered.
- R21. Seller dashboard cards show: total sales today (MVR), paid orders count, pending orders count, low-stock products count, and orders awaiting shipment.

**Swipe integration depth**

- R22. Seller dashboard includes a "Swipe Wallet" card showing available balance and pending balance from `GET /api/v1/balance`.
- R23. Seller dashboard includes a Swipe transaction history section showing recent transactions with: reference, amount, fee breakdown (`gross_amount`, `fee_amount`, `net_amount`), status, and timestamp. Data from `GET /api/v1/history`.
- R24. The app caches the Swipe OAuth2 access token and refreshes it before expiry, rather than requesting a new token for every API call.
- R25. Payment creation includes the product name and order ID in the Swipe `description` field so it appears in the buyer's Swipe app.

**Demo and presentation**

- R26. The app includes a landing page with a "before vs after" section showing the old manual flow (DM, transfer slip, manual check) alongside the new Swipe-powered flow. Includes impact messaging: time savings, fraud elimination, automatic reconciliation.
- R27. Demo data (store "Island Finds MV" with 5 sample products) is loaded via a Convex seed script. Data persists across server restarts via Convex's cloud database. A reset mechanism (admin button or re-running the seed script) allows returning to a clean demo state with deterministic data.

**Interaction states and error handling**

- R32. The checkout flow displays explicit loading states at each async boundary: (a) spinner/disabled button during stock validation after "Buy Now", (b) loading indicator during delivery form submission and Swipe payment creation, (c) skeleton/spinner while awaiting the QR code or pay-page URL from Swipe. Buttons are locked during each async operation to prevent duplicate submissions.
- R33. If the SSE connection fails to establish, drops mid-session, or times out, the app falls back to server-side polling of `GET /api/v1/payments/{id}` every 5 seconds for up to 5 minutes. If the SSE connection drops after connecting, the app attempts auto-reconnect (up to 3 attempts) before falling back to polling. The buyer sees a subtle "Checking payment status..." indicator during polling.
- R34. Delivery form validation: name (required, min 2 chars), phone (required, Maldivian 7-digit mobile number starting with 7 or 9), delivery location (required, Male' or Hulhumale'), delivery address (required, min 5 chars), delivery time preference (optional, free text). Invalid fields show inline error messages. Form submission is blocked until all required fields pass validation.

**Technical foundation**

- R28. Seller dashboard uses a simple demo authentication mechanism (e.g., a store slug or PIN code — no full auth system). The mechanism must block unauthenticated access to seller routes.
- R35. Payment amount sent to the Swipe API must be calculated server-side from the authoritative product variant price record, never from a buyer-supplied value. The server validates that the product is active, the variant exists, and stock is available before creating the payment.
- R36. Swipe client credentials (`client_id`, `client_secret`) are read exclusively from environment variables. `.env` must be listed in `.gitignore`. The cached OAuth2 access token is held in server-side memory only — not written to disk, not exposed to the browser, and not logged in server output.
- R29. All MVR amounts display with consistent formatting throughout the app (e.g., "MVR 650.00").
- R30. All timestamps in the UI display in Maldives timezone (UTC+5), not UTC.
- R31. If the Swipe mock API is unavailable, the app falls back to a local mock response mode (`SWIPE_DEMO_MODE=true`) that returns realistic payment objects without network calls.

---

## Acceptance Examples

- AE1. **Covers R6.** Given a buyer on a mobile phone (viewport < 768px), when they reach the checkout page, then a "Pay with Swipe" button linking to `payment_url` is shown instead of a QR code image.
- AE2. **Covers R6.** Given a buyer on a desktop browser, when they reach the checkout page, then a QR code image (decoded from the base64 `qr_data`) is displayed for scanning.
- AE3. **Covers R7, R8.** Given a buyer on the checkout page with an SSE connection open, when the Swipe payment transitions to COMPLETED, then the order status updates to "paid" and stock deducts without the buyer refreshing the page.
- AE4. **Covers R9.** Given a buyer on the checkout page, when the Swipe payment transitions to EXPIRED, then the page shows "Payment expired" with a "Try Again" option that creates a new payment.
- AE5. **Covers R8.** Given two buyers attempting to buy the last item of a variant simultaneously, when the first payment completes, then the second buyer's order creation is rejected with "Sorry, this item is no longer available in the selected quantity."
- AE6. **Covers R27.** Given the Next.js dev server restarts (hot reload or manual), when the app starts, then demo data is present and consistent without manual intervention.
- AE7. **Covers R31, R10.** Given `SWIPE_DEMO_MODE=true` and the Swipe mock is not running, when a buyer completes checkout, then the app creates a local mock payment and the "Simulate Payment" button completes it without network errors.
- AE8. **Covers R20.** Given a paid order, when the seller clicks "Mark as Shipped", then the order status changes to "shipped" and the order list reflects the update.

---

## Success Criteria

- A hackathon judge can watch the full demo (F4) and understand the problem, the solution, and how Swipe powers every transaction — without any manual explanation of the payment flow.
- A buyer can go from an Instagram link to a paid, confirmed order with delivery details in under 2 minutes on a mobile device.
- The seller dashboard shows payment confirmation, stock changes, wallet balance, and transaction history — all driven by Swipe API data — demonstrating deep integration beyond basic payment creation.
- The `ce-plan` agent can build a complete implementation plan from this document without needing to invent any product behavior, scope decisions, or user-facing flows.

---

## Scope Boundaries

- No real Instagram/Facebook auto-publishing (Meta API permissions are out of scope; captions are generated for manual copy-paste)
- No real Swipe credentials or production payments (mock API only)
- No buyer accounts or login (buyers are anonymous; identity is captured per-order)
- No cart or multi-product orders (single product per order; order_items schema retained for future)
- No delivery outside Male' and Hulhumale'
- No refund automation
- No multi-tenant billing or subscription
- No native mobile app (responsive web only)
- No full DM automation or chatbot
- No stock reservation/hold system (stock deducts on payment confirmation only)
- No Swipe webhook signature verification (SSE is primary; webhook architecture is deferred)
- No image upload for sellers (use image URLs/placeholders for MVP)
- No real-time seller notifications (push/websocket) — dashboard polling or refresh only
- No product search (category filtering only)

---

## Key Decisions

- **Single-product orders for MVP**: Each order is one product/variant. The order_items schema supports multi-item orders for future cart functionality, but the buyer flow is single-product. Rationale: matches social commerce pattern where buyers click through from a specific product post.
- **No stock reservation, but atomic deduction required**: Stock deducts only on payment confirmation — no pre-payment holds or reservation timers. However, stock deduction must use an atomic compare-and-decrement operation (database-level `UPDATE WHERE stock_available >= quantity` or equivalent constraint) so that concurrent payment confirmations for the last unit reject the second buyer rather than overselling. Application-level read-then-write is explicitly disallowed. Rationale: reduces complexity while satisfying AE5's concurrent-purchase guarantee.
- **Always QR type, both QR + link displayed on all viewports**: Always create QR-type payments (guaranteed mock support). Both the QR code image and a "Pay with Swipe" link button are shown on desktop and mobile. Desktop emphasizes the QR; mobile emphasizes the link button. Rationale: avoids dependency on unverified LINK payment type while giving buyers flexibility on any device.
- **SSE for real-time payment status**: Use Swipe's SSE streaming endpoint for live payment updates on the checkout page. Rationale: more impressive demo than polling, and the mock API supports it natively.
- **Delivery scoped to Male'/Hulhumale'**: Simple location selector + free-text address. Rationale: covers the majority of Maldives e-commerce delivery and keeps the form simple.
- **Convex as the database layer**: Use Convex for persistence, real-time subscriptions, and atomic mutations. Convex provides built-in real-time data sync (enabling automatic seller dashboard updates when new orders arrive), atomic transactions (satisfying AE5's concurrent-purchase guarantee), and cloud-hosted persistence (data survives server restarts). Rationale: solves persistence, atomicity, and real-time updates in one dependency. Trade-off: requires internet during demo.
- **Demo data seeding**: Seed data (store "Island Finds MV" with 5 products) is loaded via a Convex seed script. Data persists across server restarts. A reset mechanism (admin button or seed script re-run) allows returning to a clean demo state.
- **Template-based captions default, AI optional**: Social caption generation works without any API key. If an AI key (Anthropic) is configured, captions use AI. Rationale: reduces demo dependencies.
- **Mock's `POST /pay/{shortCode}/complete` for demo simulation**: Uses the Swipe mock's built-in payment completion endpoint (POST method, returns 303 redirect) rather than a custom button. Rationale: demonstrates understanding of the mock's capabilities.

---

## Dependencies / Assumptions

- The Swipe mock API (`swipe-merchants-dev`) runs locally on port 8080 and supports OAuth2, payment creation, SSE streaming, and the `/pay/{reference}/complete` demo endpoint
- Go 1.26+ is installed for building the Swipe CLI (or pre-built binary is available)
- Node.js and npm/pnpm are available for the Next.js development environment
- The hackathon demo will be presented on a machine that can run both the Swipe mock and the Next.js dev server simultaneously
- The Anthropic API key in `.env` is valid for optional AI caption generation
- Maldives sellers primarily sell fashion, gifts, accessories, and food items (reflected in demo data)
- WhatsApp is widely used in Maldives for buyer-seller communication
- The Swipe mock API key must be created with scopes: `payments:qr`, `payments:link`, `transactions:status`, `transactions:history`, `wallet:balance` (all required by R6, R7, R22, R23)
- The Swipe mock auto-transitions PENDING payments to COMPLETED after 60 seconds by default. The Simulate Payment button exists for immediate completion without waiting. To test EXPIRED or CANCELLED flows (R9), use the mock's `POST /pay/{shortCode}/expire` or `/pay/{shortCode}/cancel` endpoints

---

## Outstanding Questions

### Resolve Before Planning

(None — all product decisions resolved during brainstorm)

### Deferred to Planning

- ~~[Affects R8][Technical] Race condition~~ — resolved: atomic compare-and-decrement required (see Key Decisions)
- [Affects R24][Technical] Token caching strategy — in-memory cache with TTL, or persistent cache? How to handle token refresh timing given the mock's token TTL?
- [Affects R13][Needs research] What image placeholder strategy to use — static demo images bundled in `/public`, or a service like placeholder.com?
- ~~[Affects R6][Needs research] LINK payment type~~ — resolved: always use QR type, mobile uses mock pay page URL (see Key Decisions)
- ~~[Affects R27][Technical] Data persistence~~ — resolved: Convex database (see Key Decisions)
