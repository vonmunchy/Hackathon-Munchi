---
title: Requirements Review Changelog
date: 2026-05-20
source: docs/brainstorms/2026-05-20-swipe-social-storefront-requirements.md
review_type: ce-doc-review (7 personas)
round: 1
---

# Swipe Social Storefront — Requirements Review Changelog

## Review Summary

| Metric | Value |
|--------|-------|
| Reviewers dispatched | 7 (coherence, feasibility, product-lens, design-lens, security-lens, scope-guardian, adversarial) |
| Total findings raised | 51 |
| Findings applied | 11 |
| Findings skipped | 2 |
| FYI observations | 8 |
| Chain roots resolved | 1 (with 3 dependents) |
| Outstanding questions resolved | 3 of 5 |

---

## Applied Changes (Detailed)

### Change 1: Server-Side SSE Proxy Requirement (P0 — Critical)

**Requirement affected:** R7
**Reviewers:** feasibility, adversarial (confidence: 100)
**Finding type:** Omission

**Problem:** The Swipe mock's `/api/v1/payments/{id}/stream` SSE endpoint requires a valid OAuth2 Bearer token and checks merchant ownership (`auth.PrincipalFromContext` in `payment_stream.go` lines 41-56). The original R7 said "the checkout page opens an SSE connection to the Swipe payment stream" — implying a direct browser-to-Swipe connection. The browser's `EventSource` API cannot set Authorization headers, so this would fail with a 401 error. The entire real-time payment confirmation demo (the hackathon centerpiece) would break.

**Solution:** R7 rewritten to require all Swipe API communication to be server-to-server. The Next.js backend proxies the SSE stream via a server-side API route (e.g., `/api/checkout/[orderId]/stream`) that authenticates to Swipe with the cached merchant OAuth token and re-emits events to the browser client.

**Before:**
```
R7. The checkout page opens an SSE connection to the Swipe payment stream 
(`/api/v1/payments/{id}/stream`) and updates the payment status in real-time 
without page refresh.
```

**After:**
```
R7. The checkout page receives real-time payment status updates without page 
refresh. All Swipe API communication is server-to-server — the Next.js backend 
proxies the Swipe SSE stream via a server-side API route (e.g., 
`/api/checkout/[orderId]/stream`) that authenticates to Swipe with the cached 
merchant OAuth token and re-emits events to the browser client. The buyer's 
browser never connects directly to the Swipe API.
```

**Why this works:** Server-side proxy keeps the OAuth2 merchant credentials secure (never exposed to the browser), satisfies the Swipe API's authentication requirement, and provides a single architecture pattern for all Swipe API interactions.

---

### Change 2: Correct Mock Pay-Page Endpoint (safe_auto fix)

**Requirement affected:** R10, Key Decisions
**Reviewer:** feasibility (confidence: 100)
**Finding type:** Error

**Problem:** R10 referenced `/pay/{reference}/complete` but the actual mock route is `POST /pay/{shortCode}/complete` (pay_page.go line 53). The endpoint also returns a 303 redirect, not JSON — the simulate button needs to handle this correctly.

**Solution:** Updated R10 and Key Decisions to reference the correct endpoint (`POST /pay/{shortCode}/complete`) with the 303 redirect behavior noted.

**Before:**
```
R10. A "Simulate Payment" fallback button is available on the checkout page 
that calls the mock's `/pay/{reference}/complete` endpoint, for demo purposes 
and as a fallback if SSE fails.
```

**After:**
```
R10. A "Simulate Payment" fallback button is available on the checkout page 
that fires a server-side POST to the mock's `/pay/{shortCode}/complete` endpoint 
(where shortCode is the payment's `short_code`/`reference` value). The endpoint 
returns a 303 redirect, not JSON — the button should fire-and-forget. For demo 
purposes and as a fallback if SSE fails.
```

---

### Change 3: Required Swipe OAuth2 Scopes (safe_auto fix)

**Section affected:** Dependencies / Assumptions
**Reviewer:** adversarial (confidence: 100)
**Finding type:** Omission

**Problem:** The Swipe API requires different OAuth2 scopes for different operations (`payments:qr` for QR, `wallet:balance` for balance, `transactions:history` for history, `transactions:status` for SSE). The requirements didn't specify which scopes to request. If the planner creates a key with only `payments:qr`, wallet balance (R22), transaction history (R23), and SSE (R7) would all fail with 403 Forbidden.

**Solution:** Added explicit scope requirements to Dependencies section.

**Added:**
```
- The Swipe mock API key must be created with scopes: `payments:qr`, 
  `payments:link`, `transactions:status`, `transactions:history`, 
  `wallet:balance` (all required by R6, R7, R22, R23)
```

---

### Change 4: Mock Auto-Completion TTL Behavior (safe_auto fix)

**Section affected:** Dependencies / Assumptions
**Reviewer:** feasibility (confidence: 100)
**Finding type:** Omission

**Problem:** The mock automatically transitions PENDING payments to COMPLETED after 60 seconds (`PaymentTransitionTTL` in `payment_create.go` line 47). Without knowing this, implementers would design unnecessary payment timeout and retry logic, and would not know how to test EXPIRED/CANCELLED flows.

**Solution:** Added TTL behavior to Dependencies section.

**Added:**
```
- The Swipe mock auto-transitions PENDING payments to COMPLETED after 60 seconds 
  by default. The Simulate Payment button exists for immediate completion without 
  waiting. To test EXPIRED or CANCELLED flows (R9), use the mock's 
  `POST /pay/{shortCode}/expire` or `/pay/{shortCode}/cancel` endpoints
```

---

### Change 5: R12/R16 Dependency Made Explicit (safe_auto fix)

**Requirement affected:** R12
**Reviewer:** scope-guardian (confidence: 100)
**Finding type:** Omission

**Problem:** R12 requires "Message Seller" deep links to WhatsApp and Instagram DM, but these depend on seller contact info from R16 (store profile). The dependency was implicit — if R16 isn't implemented, R12's deep links would be broken or hardcoded.

**Solution:** Added explicit dependency note to R12.

**Added to R12:**
```
(Depends on R16: reads seller WhatsApp number and Instagram handle from the 
store profile. If not configured, deep links fall back to a placeholder or 
are omitted.)
```

---

### Change 6: Atomic Stock Deduction + Race Condition Resolution (P1)

**Sections affected:** Key Decisions, Outstanding Questions
**Reviewers:** product-lens, coherence, scope-guardian, adversarial (confidence: 100)
**Finding type:** Error (contradiction)

**Problem:** AE5 explicitly requires rejecting the second buyer when two attempt to buy the last unit simultaneously. But Key Decisions said "No stock reservation" and the race condition mechanism was deferred to planning. These are contradictory commitments — you cannot guarantee AE5 without atomic operations, yet the doc told planners not to build reservation logic.

**Solution:** Clarified that "no reservation" means no pre-payment holds/timers, but stock deduction must be atomic (compare-and-decrement). Application-level read-then-write is explicitly disallowed. Moved from "Deferred to Planning" to Key Decisions.

**Before (Key Decision):**
```
No stock reservation: Stock deducts only on payment confirmation. 
No temporary holds or reservation timers.
```

**After (Key Decision):**
```
No stock reservation, but atomic deduction required: Stock deducts only on 
payment confirmation — no pre-payment holds or reservation timers. However, 
stock deduction must use an atomic compare-and-decrement operation (database-level 
UPDATE WHERE stock_available >= quantity or equivalent constraint) so that 
concurrent payment confirmations for the last unit reject the second buyer rather 
than overselling. Application-level read-then-write is explicitly disallowed.
```

**Outstanding Question resolved:** Race condition handling — now a Key Decision.

---

### Change 7: Convex Database for Persistence (P1)

**Sections affected:** Key Decisions, R27, Outstanding Questions
**Reviewers:** scope-guardian, adversarial (multiple findings)
**Finding type:** Error (AE6/R27 contradiction)

**Problem:** R27 said data resets on cold start. AE6 said data is "present and consistent" after hot reload. Next.js hot reload wipes in-memory module-scope variables, making these contradictory. Additionally, the atomic stock deduction requirement (Change 6) is very hard to implement with in-memory data.

**Solution:** Adopted Convex as the database layer. Convex provides:
- Built-in real-time subscriptions (automatic seller dashboard updates)
- Atomic transactions via mutations (satisfies AE5)
- Cloud-hosted persistence (data survives restarts)
- First-class React/Next.js bindings

Updated R27 to reference Convex seed script instead of auto-seeding on startup. Added reset mechanism for demo purposes.

**Before (Key Decision):**
```
Demo data auto-seeds on startup: In-memory data resets to a clean state on 
every server start.
```

**After (Key Decisions):**
```
Convex as the database layer: Use Convex for persistence, real-time subscriptions, 
and atomic mutations. [...]

Demo data seeding: Seed data (store "Island Finds MV" with 5 products) is loaded 
via a Convex seed script. Data persists across server restarts. A reset mechanism 
(admin button or seed script re-run) allows returning to a clean demo state.
```

**Outstanding Question resolved:** Data persistence strategy — Convex.

---

### Change 8: QR + Link on All Viewports (P1)

**Sections affected:** R6, Key Decisions, Outstanding Questions
**Reviewers:** product-lens (confidence: 75)
**Finding type:** Error (deferred question blocking mobile checkout)

**Problem:** R6 mandated LINK payment type for mobile, but it was unknown if the mock supports LINK. The mock may only support QR. Given that buyers come from Instagram (mobile), mobile checkout is the primary path.

**Solution:** Always create QR-type payments (guaranteed mock support). Display **both** QR code and pay-page link on all viewports:
- Desktop: QR prominently displayed, link as secondary
- Mobile: Link button prominently displayed, smaller QR below

**Outstanding Question resolved:** LINK payment type — bypassed by using QR type with mock pay-page URL.

---

### Change 9: Loading States Requirement (P1)

**Requirement added:** R32
**Reviewer:** design-lens (confidence: 100)
**Finding type:** Omission

**Problem:** Between clicking "Buy Now" and seeing the payment QR/link, there are multiple async operations (stock check, delivery form submission, payment creation) with no defined intermediate UI state. Buyers could click again, creating duplicate payments.

**Added:**
```
R32. The checkout flow displays explicit loading states at each async boundary: 
(a) spinner/disabled button during stock validation after "Buy Now", 
(b) loading indicator during delivery form submission and Swipe payment creation, 
(c) skeleton/spinner while awaiting the QR code or pay-page URL from Swipe. 
Buttons are locked during each async operation to prevent duplicate submissions.
```

---

### Change 10: SSE Failure Fallback Requirement (P1)

**Requirement added:** R33
**Reviewers:** design-lens, adversarial (confidence: 100)
**Finding type:** Omission

**Problem:** R7 opens SSE but doesn't define what happens on connection failure, drop, or timeout. Combined with "no webhook" scope boundary, a dropped SSE connection means the payment completes in Swipe but the app never confirms the order.

**Added:**
```
R33. If the SSE connection fails to establish, drops mid-session, or times out, 
the app falls back to server-side polling of GET /api/v1/payments/{id} every 
5 seconds for up to 5 minutes. If the SSE connection drops after connecting, 
the app attempts auto-reconnect (up to 3 attempts) before falling back to 
polling. The buyer sees a subtle "Checking payment status..." indicator 
during polling.
```

---

### Change 11: Delivery Form Validation Requirement (P1)

**Requirement added:** R34
**Reviewer:** design-lens (confidence: 100)
**Finding type:** Omission

**Problem:** R5 lists delivery form fields but specifies no validation rules, error display, or required/optional distinction. Implementation would invent validation inconsistently.

**Added:**
```
R34. Delivery form validation: name (required, min 2 chars), phone (required, 
Maldivian 7-digit mobile number starting with 7 or 9), delivery location 
(required, Male' or Hulhumale'), delivery address (required, min 5 chars), 
delivery time preference (optional, free text). Invalid fields show inline 
error messages. Form submission is blocked until all required fields pass 
validation.
```

---

### Change 12: Server-Side Price Validation Requirement (P1)

**Requirement added:** R35
**Reviewer:** security-lens (confidence: 75)
**Finding type:** Omission

**Problem:** If the payment amount sent to Swipe is derived from a buyer-supplied value rather than looked up server-side, a buyer could manipulate the price. Payment confirmation triggers stock deduction, so an under-priced payment could complete and consume inventory.

**Added:**
```
R35. Payment amount sent to the Swipe API must be calculated server-side from 
the authoritative product variant price record, never from a buyer-supplied 
value. The server validates that the product is active, the variant exists, 
and stock is available before creating the payment.
```

---

### Change 13: Secrets Management Requirement (P1)

**Requirement added:** R36
**Reviewer:** security-lens (confidence: 100)
**Finding type:** Omission

**Problem:** R24 caches the OAuth token, but no requirement specified where credentials are stored, that .env is gitignored, or that tokens don't appear in logs.

**Added:**
```
R36. Swipe client credentials (client_id, client_secret) are read exclusively 
from environment variables. .env must be listed in .gitignore. The cached 
OAuth2 access token is held in server-side memory only — not written to disk, 
not exposed to the browser, and not logged in server output.
```

---

## Skipped Changes

### Auth Hardening (P1, security-lens)
**Reason:** User chose to defer auth hardening (PIN length, rate limiting, session mechanism) as non-critical for hackathon MVP. The demo auth (R28) remains as specified.

### Buyer PII Protection (P1, security-lens)
**Reason:** User chose to defer PII protection requirements (encryption, logging, retention) as non-critical for hackathon MVP. Buyer data (name, phone, address) has no explicit protection requirement beyond HTTPS.

---

## FYI Observations (Not Actioned — Informational Only)

1. **R5 delivery location scope** — could be more explicitly bound to scope boundary (coherence)
2. **R3/R4 social sharing vs OG tags** — two requirements could be confused as overlapping (coherence)
3. **R18 timezone redundancy** — R18 repeats timezone spec that R30 already covers globally (coherence)
4. **R22/R23 serve judges, not sellers** — wallet/history features address hackathon judging criteria, not stated seller problems (product-lens)
5. **Dashboard zero state** — R21 cards have no defined empty/zero state for demo startup (design-lens)
6. **SWIPE_DEMO_MODE default** — should default to false with startup warning (security-lens)
7. **Do-nothing baseline** — the before/after section (R26) doesn't quantify the manual flow's pain (adversarial)
8. **order_items dead weight** — join table is unused for single-product MVP; flat order record is simpler (adversarial)

---

## Resolved Outstanding Questions

| Original Question | Resolution |
|---|---|
| Race condition for simultaneous last-unit purchases | Atomic compare-and-decrement required (Key Decision) |
| LINK payment type mock support | Bypassed: always use QR type, display pay-page URL as link |
| In-memory vs persistent data storage | Convex database adopted |

## Remaining Outstanding Questions

| Question | Status |
|---|---|
| Token caching strategy (R24) | Deferred to Planning |
| Image placeholder strategy (R13) | Deferred to Planning |
