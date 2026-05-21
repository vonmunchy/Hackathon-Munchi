# SwiftStore

**The Social Commerce Platform for Maldivian Sellers**

> **Proof of Concept** — This project is a hackathon submission demonstrating the viability of a unified social commerce + P2P exchange platform for the Maldives. While fully functional end-to-end in demo mode, it uses simulated blockchain wallets and mock payment webhooks. The architecture and UX are production-grade; the integrations are ready to be connected to real infrastructure.

SwiftStore is a full-stack social commerce platform that empowers Maldivian small businesses to sell products through their social media presence, accept payments via Swipe, and trade USDT through a built-in P2P exchange — all from a single dashboard.

---

## Live Demo

- **Production**: [mvswiftstore.vercel.app](https://mvswiftstore.vercel.app)
- **Demo Store**: Visit `/seller/login` → use slug `island-crafts` with PIN `1234`
- **Marketplace**: Browse all products at `/marketplace`
- **P2P Exchange**: Trade USDT at `/exchange`
- **Admin Panel**: `/backstage` (passphrase: `swiftstore-admin-2024`)

---

## Problem Statement

In the Maldives, thousands of small sellers operate exclusively through Instagram and Facebook. They face:

- **No unified storefront** — products scattered across social posts
- **Manual order management** — tracking orders via DMs and spreadsheets
- **Limited payment options** — no integrated digital payment flow
- **No crypto on-ramp** — growing demand for USDT but no local P2P platform

---

## Solution

SwiftStore provides:

1. **Instant Storefronts** — Sellers get a shareable product page (`/shop/your-store`) in minutes
2. **Swipe Payment Integration** — QR-based checkout with real-time payment confirmation
3. **P2P USDT Exchange** — Sellers list USDT, buyers pay in MVR via Swipe, trustless escrow
4. **AI-Powered Social Integration** — Trace seller's social posts, auto-generate inventory, publish directly
5. **Multi-tenant Dashboard** — Each seller manages products, orders, payments, and exchange listings

---

## AI-Powered Inventory Generation (Vision)

One of SwiftStore's key differentiators is how it bridges the gap between social media posts and a real product catalog:

### How It Works

```
Seller connects Instagram/Facebook → SwiftStore reads their posts via Meta Graph API
  → AI (Claude) analyzes post images + captions → Extracts product name, description,
    price, category, variants → Auto-generates inventory draft
  → Seller reviews, edits, confirms → Products go live on their storefront
```

1. **Social Post Tracing**: When a seller connects their Instagram or Facebook account via OAuth, SwiftStore accesses their recent posts through the Meta Graph API
2. **AI Analysis**: Each post's image and caption are analyzed by Claude (Anthropic API) to extract structured product data — name, description, estimated price, category, and potential variants (sizes, colors)
3. **Draft Inventory**: The AI-generated products are presented to the seller as editable drafts, pre-filled with information extracted from their posts
4. **Seller Confirmation**: The seller reviews each product, adjusts prices, adds variants, uploads additional images, and confirms — turning social posts into a real catalog
5. **Caption Generation**: When sellers want to promote products back on social media, AI generates optimized captions with hashtags tailored for the Maldivian market

This creates a flywheel: **Social posts → AI-generated inventory → Storefront → Sales → New social posts with AI captions → More inventory**.

### Current Implementation

For this hackathon proof of concept:
- The Meta Graph API integration is **functional** for reading posts and publishing content
- AI caption generation is **implemented** using Claude API (`/api/social/caption`)
- The full inventory auto-generation pipeline is **architected but not fully wired** — the individual pieces (Meta API read, AI analysis, product creation) all work independently
- Sellers can currently import demo products or manually create products, with AI-assisted social publishing

### Meta API Limitation

> **Important for judges**: Meta requires a **4-6 week App Review process** before third-party Facebook/Instagram accounts can authorize with our app. During this review period, only the developer's own accounts can be used for OAuth login and social publishing. This is a standard Meta platform requirement, not a limitation of our implementation.
>
> For the hackathon demo, we use the developer's own Meta account to demonstrate the full OAuth flow, post reading, and social publishing. The code is written to support any authorized account — once Meta approves the app, any seller can connect their pages.

### Meta Permissions Requested
- `pages_show_list` — List seller's Facebook Pages
- `pages_read_engagement` — Read posts and engagement data
- `pages_manage_posts` — Publish product posts
- `instagram_basic` — Read Instagram profile and media
- `instagram_content_publish` — Publish to Instagram

---

## Key Features

### For Sellers
| Feature | Description |
|---------|-------------|
| **Multi-step Onboarding** | Choose path (marketplace/crypto/both), connect Swipe, import products |
| **Product Management** | Multi-variant products (size, color), image upload, stock tracking |
| **Order Management** | Real-time order status (pending → paid → shipped → delivered) |
| **Swipe Payments** | Automatic QR generation, payment status polling, settlement |
| **P2P Exchange** | Create USDT listings, set rates, partial fills, simulated TRC20 wallets |
| **Social Integration** | Facebook/Instagram OAuth, AI caption generation, direct publishing, post tracing |
| **AI Inventory** | Analyze social posts to auto-generate product catalog drafts for seller review |
| **Store Settings** | Custom description, social links, WhatsApp, logo upload |

### For Buyers
| Feature | Description |
|---------|-------------|
| **Marketplace** | Browse all products, filter by category/seller, search |
| **Store Pages** | Visit individual seller storefronts via shareable links |
| **Instant Checkout** | Select variant → enter delivery info → scan QR → order confirmed |
| **P2P Exchange** | Browse USDT listings sorted by rate, reserve, pay via Swipe |
| **Post-Purchase** | Contact seller via WhatsApp/Instagram/Messenger |

### Admin (Backstage)
- Passphrase-protected admin panel
- View all registered sellers with onboarding/credential status
- Reset seller onboarding for re-registration

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 16, React 19, TypeScript |
| **Styling** | Tailwind CSS 4, Framer Motion |
| **Backend** | Convex (real-time database, serverless functions) |
| **Payments** | Swipe API (OAuth2, QR payments, webhooks) |
| **Auth** | Session-based with PIN login + Meta OAuth (Facebook/Instagram) |
| **Crypto** | TRC20 USDT (simulated wallets for hackathon) |
| **AI** | Anthropic Claude API (caption generation, inventory extraction) |
| **Deployment** | Vercel (frontend) + Convex Cloud (backend) |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    SWIFTSTORE                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────┐  │
│  │  Buyer   │    │  Seller  │    │     Admin        │  │
│  │  Views   │    │Dashboard │    │   (Backstage)    │  │
│  └────┬─────┘    └────┬─────┘    └────────┬─────────┘  │
│       │               │                    │            │
│  ┌────┴───────────────┴────────────────────┴─────────┐  │
│  │           Next.js 16 (App Router)                  │  │
│  │     Mobile Shell ←→ Desktop Shell (auto-detect)   │  │
│  └────────────────────────┬──────────────────────────┘  │
│                           │                             │
│  ┌────────────────────────┴──────────────────────────┐  │
│  │              API Routes (Next.js)                  │  │
│  │  /api/checkout  /api/swipe  /api/admin            │  │
│  └──────┬─────────────┬──────────────┬───────────────┘  │
│         │             │              │                  │
│  ┌──────┴──┐   ┌──────┴──────┐  ┌───┴────────────┐    │
│  │ Convex  │   │  Swipe API  │  │  Meta Graph    │    │
│  │ Backend │   │  (Payments) │  │  API (Social)  │    │
│  └─────────┘   └─────────────┘  └───┬────────────┘    │
│                                      │                  │
│                              ┌───────┴────────────┐    │
│                              │  Claude AI (Anthropic)│   │
│                              │  Captions + Inventory │   │
│                              └────────────────────┘    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Data Flow: Checkout
```
Buyer selects product → Enters delivery info → Creates order (Convex)
  → Generates Swipe payment (QR + short code) → Buyer scans QR
  → Swipe confirms payment → Order marked as PAID → Stock decremented
  → Seller notified → Buyer sees success page with seller contact
```

### Data Flow: AI Inventory Generation
```
Seller connects Instagram → Meta Graph API fetches recent posts
  → Claude AI analyzes images + captions → Extracts product data
  → Draft products created in Convex → Seller reviews & confirms
  → Products go live on storefront → AI generates social captions
  → Seller publishes back to Instagram/Facebook → Cycle repeats
```

### Data Flow: P2P Exchange
```
Seller creates USDT listing (amount + rate) → Simulates TRC20 deposit
  → Listing goes ACTIVE → Buyer reserves amount (5-min window)
  → Buyer pays MVR via Swipe → Payment confirmed
  → USDT "transferred" to buyer wallet → Transaction recorded
```

---

## Responsive Design

SwiftStore uses a **dual-layout system** (not just responsive breakpoints):

- **Mobile**: App-like experience with bottom tab bar, full-width cards, touch-optimized
- **Desktop**: Sidebar navigation for sellers, multi-column grids, hover states

Device detection happens at the layout level via `useIsMobile()` hook, routing to entirely different shell components for optimal UX on each platform.

---

## Database Schema

```
stores              → Seller profiles, credentials, social links
products            → Multi-variant product catalog
productVariants     → Size/color/stock per variant
orders              → Full order lifecycle with Swipe payment data
sessions            → 24-hour auth tokens (multi-tenant isolation)
exchangeListings    → USDT sell orders (rate, balance, status)
exchangeTransactions → Completed P2P trades
exchangeReservations → Time-locked purchase intents (5-min expiry)
```

---

## Security & Multi-Tenancy

- **Session-based authentication** with 24-hour expiry
- **Multi-tenant isolation**: Every seller mutation validates store ownership
- **Access tokens**: Buyers access orders via random UUID tokens (no account needed)
- **Admin passphrase gate**: Backstage protected by environment variable
- **Input validation**: Convex schemas enforce types; TRC20 address format validated
- **No secrets in repo**: All credentials via environment variables (`.env.local`)

---

## Getting Started

### Prerequisites
- Node.js 18+
- npm
- Convex account ([convex.dev](https://convex.dev))
- Swipe merchant account (optional — demo mode available)

### Installation

```bash
# Clone the repository
git clone https://github.com/vonmunchy/Hackathon-Munchi.git
cd Hackathon-Munchi/swipe-social-storefront

# Install dependencies
npm install

# Set up Convex
npx convex dev    # Creates .env.local with CONVEX_DEPLOYMENT and NEXT_PUBLIC_CONVEX_URL

# Configure environment variables
cp .env.local.example .env.local
# Edit .env.local with your credentials (see Environment Variables below)

# Seed demo data
npm run seed

# Start development server
npm run dev
```

### Environment Variables

Create `swipe-social-storefront/.env.local`:

```env
# Convex (auto-populated by `npx convex dev`)
CONVEX_DEPLOYMENT=dev:your-deployment-name
NEXT_PUBLIC_CONVEX_URL=https://your-deployment.convex.cloud

# Swipe Payment API
SWIPE_API_BASE_URL=http://127.0.0.1:8080   # or production URL
SWIPE_CLIENT_ID=your_client_id
SWIPE_CLIENT_SECRET=your_client_secret
SWIPE_DEMO_MODE=true                        # Set false for real payments

# Meta API (optional — for social publishing)
META_APP_ID=your_app_id
META_PAGE_ID=your_page_id
META_PAGE_ACCESS_TOKEN=your_token
META_APP_SECRET=your_secret

# Demo Controls
NEXT_PUBLIC_SHOW_DEMO_CONTROLS=true         # Shows simulate buttons on checkout

# Admin
ADMIN_PASSPHRASE=your-admin-passphrase
```

---

## Demo Mode

For hackathon judging, the app runs in **demo mode** (`SWIPE_DEMO_MODE=true`):

- Payment QR codes are generated with mock data
- "Simulate Payment" button appears on checkout to instantly confirm orders
- Exchange deposits are simulated (no real blockchain interaction)
- All core flows (onboarding → listing → checkout → confirmation) work end-to-end

### Demo Credentials

| Store | Slug | PIN |
|-------|------|-----|
| Island Crafts | `island-crafts` | `1234` |

---

## Project Structure

```
SwiftStore/
├── swipe-social-storefront/        # Main web application
│   ├── app/                        # Next.js App Router pages
│   │   ├── seller/                 # Seller dashboard routes
│   │   ├── marketplace/            # Public marketplace
│   │   ├── exchange/               # P2P USDT exchange
│   │   ├── shop/[storeSlug]/       # Individual store pages
│   │   ├── checkout/[orderId]/     # Payment flow
│   │   ├── backstage/              # Admin panel
│   │   └── api/                    # API routes (Swipe, checkout, admin)
│   ├── components/                 # UI components
│   │   ├── buyer/                  # Buyer-facing components
│   │   ├── seller/                 # Seller dashboard components
│   │   ├── mobile/                 # Mobile-specific (tab bar, bottom sheet)
│   │   ├── desktop/                # Desktop-specific (sidebar, nav)
│   │   └── shells/                 # Layout shells (mobile/desktop router)
│   ├── convex/                     # Backend functions & schema
│   │   ├── schema.ts              # Database schema
│   │   ├── products.ts            # Product CRUD
│   │   ├── orders.ts              # Order lifecycle
│   │   ├── exchange.ts            # P2P exchange logic
│   │   ├── stores.ts             # Store management
│   │   ├── auth.ts               # Multi-tenant auth
│   │   └── marketplace.ts        # Public product aggregation
│   ├── lib/                        # Utilities
│   │   ├── swipe-client.ts        # Swipe API integration
│   │   ├── swipe-demo.ts          # Demo mode mock
│   │   └── use-device.ts          # Mobile/desktop detection
│   └── public/                     # Static assets
│       └── demo-products/          # Sample product images
├── swipe-merchants-dev/            # Swipe mock API server (Go)
└── docs/                           # Design documentation
```

---

## Hackathon Context

**Event**: Swipe Hackathon 2025  
**Track**: Social Commerce + Crypto  
**Team**: Munchi  

### Proof of Concept Scope

This is a working proof of concept. The following are simulated for the hackathon:
- TRC20 wallet addresses and transaction hashes (randomly generated, not on-chain)
- Swipe payment webhooks (demo mode with simulate buttons)
- Per-store Swipe credential isolation (uses global credentials in demo)

Everything else — the UI, database, order lifecycle, multi-tenant auth, exchange escrow logic — is real and functional.

### What Makes This Special

1. **Real-world problem**: Maldivian sellers actually operate this way today — via social DMs
2. **Full integration**: End-to-end from product listing to payment settlement
3. **P2P Exchange**: Novel feature — no existing local USDT↔MVR platform
4. **Dual-platform UX**: True mobile-first AND desktop experience (not just responsive)
5. **Production-ready architecture**: Multi-tenant, session-based auth, real-time Convex backend
6. **Demo mode**: Fully functional without real payment infrastructure

---

## Screenshots

| Mobile Marketplace | Desktop Seller Dashboard | P2P Exchange |
|:--:|:--:|:--:|
| Browse & filter products | Manage orders & inventory | Buy/sell USDT |

| Checkout Flow | Onboarding | Admin Backstage |
|:--:|:--:|:--:|
| QR scan → instant confirmation | Multi-step guided setup | Seller management |

---

## API Endpoints

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/checkout/create` | Create order + Swipe payment |
| GET | `/api/checkout/[orderId]/stream` | Poll payment status |
| POST | `/api/swipe/payments/create` | Create standalone payment |
| GET | `/api/swipe/payments/[id]/status` | Check payment status |
| POST | `/api/swipe/test-credentials` | Validate Swipe creds |
| POST | `/api/swipe/payments/simulate` | Demo: simulate payment |
| POST | `/api/admin/stores` | List all stores (admin) |
| POST | `/api/admin/reset-store` | Reset store (admin) |

---

## Future Roadmap

### Post-Hackathon (Ready to Build)
- [ ] Complete AI inventory auto-generation pipeline (pieces exist, need full wiring)
- [ ] Meta App Review approval (submitted, awaiting 4-6 week review)
- [ ] Real TRC20 blockchain integration (Tron network)
- [ ] Swipe webhook for instant payment confirmation (vs polling)
- [ ] Per-store Swipe credential isolation (architecture exists)

### Medium Term
- [ ] Seller analytics dashboard (revenue, conversion rates, top products)
- [ ] Multi-language support (Dhivehi + English)
- [ ] Push notifications for order updates
- [ ] Inventory alerts and auto-restock suggestions
- [ ] Rating & review system for sellers
- [ ] Bulk product import from CSV/spreadsheet

---

## Built With

- [Next.js](https://nextjs.org/) — React framework
- [Convex](https://convex.dev/) — Real-time backend
- [Tailwind CSS](https://tailwindcss.com/) — Utility-first styling
- [Framer Motion](https://www.framer.com/motion/) — Animations
- [Swipe](https://swipe.mv/) — Payment processing
- [Vercel](https://vercel.com/) — Deployment

---

## License

MIT

---

## Team

**Munchi** — Built for Swipe Hackathon 2025
