# Swipe Social Storefront — Design System

*Last updated: 2026-05-21*

## Philosophy: Premium Fintech

Rich amethyst violet meets crimson ruby on clean white. A fusion of Bank of Maldives' institutional trust and Swipe's digital-first innovation. Products float in generous space, typography is tight and refined, interactions feel native. Two completely separate layout shells serve the same data through different bones.

---

## Design Tokens

### Color Palette

```
/* Amethyst — primary brand (Swipe violet heritage) */
--amethyst-50:  #faf5ff
--amethyst-100: #f3e8ff
--amethyst-200: #e9d5ff
--amethyst-300: #d8b4fe
--amethyst-400: #c084fc
--amethyst-500: #a855f7    /* primary — Swipe violet */
--amethyst-600: #9333ea
--amethyst-700: #7e22ce
--amethyst-800: #6b21a8
--amethyst-900: #581c87

/* Ruby — accent (BML crimson heritage) */
--ruby-50:  #fff1f2
--ruby-100: #ffe4e6
--ruby-200: #fecdd3
--ruby-300: #fda4af
--ruby-400: #fb7185
--ruby-500: #f43f5e    /* accent — BML crimson */
--ruby-600: #e11d48
--ruby-700: #be123c

/* Slate — neutral surfaces (cool, premium) */
--slate-50:  #f8fafc
--slate-100: #f1f5f9
--slate-200: #e2e8f0
--slate-300: #cbd5e1
--slate-400: #94a3b8
--slate-500: #64748b
--slate-600: #475569
--slate-700: #334155
--slate-800: #1e293b
--slate-900: #0f172a

/* Semantic */
--surface:     #ffffff
--surface-alt: var(--slate-50)
--text:        var(--slate-900)
--text-muted:  var(--slate-500)
--border:      var(--slate-200)
--success:     #059669
--warning:     #d97706
--error:       #dc2626
```

### Typography

```
/* Font stack */
--font-display: 'Inter', sans-serif               /* headlines — clean, premium fintech */
--font-body:    'DM Sans', sans-serif              /* body — geometric, readable */
--font-mono:    'JetBrains Mono', monospace        /* prices, codes, references */

/* Scale */
--text-xs:   0.75rem / 1rem       /* 12px — badges, meta */
--text-sm:   0.875rem / 1.25rem   /* 14px — secondary text */
--text-base: 1rem / 1.5rem        /* 16px — body */
--text-lg:   1.125rem / 1.75rem   /* 18px — large body */
--text-xl:   1.25rem / 1.75rem    /* 20px — section titles */
--text-2xl:  1.5rem / 2rem        /* 24px — page titles */
--text-3xl:  1.875rem / 2.25rem   /* 30px — hero mobile */
--text-4xl:  2.25rem / 2.5rem     /* 36px — hero desktop */

/* Weight */
--font-light:    300   /* body text — whisper */
--font-regular:  400   /* default */
--font-medium:   500   /* labels, buttons */
--font-semibold: 600   /* headlines */
--font-bold:     700   /* hero only */
```

### Spacing

```
/* Generous — abundance, not emptiness */
--space-1:  0.25rem   /* 4px */
--space-2:  0.5rem    /* 8px */
--space-3:  0.75rem   /* 12px */
--space-4:  1rem      /* 16px */
--space-5:  1.25rem   /* 20px */
--space-6:  1.5rem    /* 24px */
--space-8:  2rem      /* 32px */
--space-10: 2.5rem    /* 40px */
--space-12: 3rem      /* 48px */
--space-16: 4rem      /* 64px */
--space-20: 5rem      /* 80px */
```

### Radius

```
/* Refined, premium curves */
--radius-sm:   0.375rem   /* 6px — inputs, badges */
--radius-md:   0.75rem    /* 12px — cards, buttons */
--radius-lg:   1rem        /* 16px — modals, sheets */
--radius-xl:   1.5rem      /* 24px — product images */
--radius-full: 9999px      /* pills, avatars */
```

### Shadows

```
/* Clean, elevated, subtle cool tint */
--shadow-sm:  0 1px 2px rgba(15, 23, 42, 0.04)
--shadow-md:  0 4px 12px rgba(15, 23, 42, 0.07)
--shadow-lg:  0 8px 24px rgba(15, 23, 42, 0.09)
--shadow-xl:  0 16px 48px rgba(15, 23, 42, 0.12)
--shadow-sheet: 0 -4px 24px rgba(15, 23, 42, 0.14)  /* bottom sheet */
```

### Motion

```
/* Unhurried confidence */
--ease-out:    cubic-bezier(0.16, 1, 0.3, 1)       /* exits, slides */
--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1)   /* bouncy entrances */
--ease-smooth: cubic-bezier(0.4, 0, 0.2, 1)        /* standard */

--duration-fast:   150ms
--duration-normal: 250ms
--duration-slow:   400ms
--duration-sheet:  350ms   /* bottom sheet rise/fall */
```

---

## Layout Architecture

### Detection Strategy

```tsx
// lib/use-device.ts — JS-based, not CSS media queries
// Runs on mount, returns stable value for the session

'use client'
import { useState, useEffect } from 'react'

export function useIsMobile(): boolean | null {
  const [isMobile, setIsMobile] = useState<boolean | null>(null)
  
  useEffect(() => {
    // Check actual viewport width at mount time
    const check = () => window.innerWidth < 768
    setIsMobile(check())
    // Do NOT add resize listener — layout stays fixed for the session
    // If user rotates tablet, they get the layout that loaded
  }, [])
  
  return isMobile  // null = SSR/loading, true = mobile, false = desktop
}
```

### Root Layout Router

```tsx
// app/layout.tsx — serves the correct shell
export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <ConvexClientProvider>
          <LayoutRouter>{children}</LayoutRouter>
        </ConvexClientProvider>
      </body>
    </html>
  )
}

// components/shared/layout-router.tsx
'use client'
export function LayoutRouter({ children }) {
  const isMobile = useIsMobile()
  if (isMobile === null) return <SplashLoader />  // brief loading
  if (isMobile) return <MobileShell>{children}</MobileShell>
  return <DesktopShell>{children}</DesktopShell>
}
```

---

## Mobile Layout (< 768px) — Native App Feel

### Shell Structure

```
┌─────────────────────────┐
│     Status Bar Area      │  ← safe-area-inset-top
├─────────────────────────┤
│                         │
│                         │
│     Full-Screen Page    │  ← scrollable content area
│     Content             │     100vh - nav - safe areas
│                         │
│                         │
├─────────────────────────┤
│  🏠   🛍️   📦   👤    │  ← Bottom tab bar (seller)
└─────────────────────────┘     OR no nav bar (buyer — back arrow only)
```

### Mobile Navigation

**Buyer pages** (storefront, product, checkout, success):
- NO bottom tab bar — clean, full-screen experience
- Back arrow top-left (← ) for navigation
- Store name centered in top bar
- Page transitions: slide in from right, slide out to left

**Seller pages** (dashboard, products, orders, social, settings):
- Bottom tab bar with 5 icons: Dashboard, Products, Orders, Social, Settings
- Active tab uses amethyst-600 violet fill, inactive uses slate-400
- Tab bar height: 56px + safe-area-inset-bottom
- Tab bar background: white with top border (slate-200)
- Page content scrolls independently above tab bar

### Mobile Page Patterns

**Product Grid (Buyer Storefront)**
```
┌─────────────────────────┐
│  ← Island Finds MV      │  top bar
├─────────────────────────┤
│ All │Fashion│Gifts│Food  │  category tabs — horizontal scroll
├────────────┬────────────┤
│            │            │
│  Product   │  Product   │  2-column grid
│  Card      │  Card      │  square images
│            │            │  name + price below
├────────────┼────────────┤
│            │            │
│  Product   │  Product   │
│  Card      │  Card      │
│            │            │
└────────────┴────────────┘
```

- Cards: 2-column grid, 8px gap
- Image: square aspect ratio, rounded-xl corners
- Below image: product name (medium weight), price (mono font, amethyst-600)
- Stock badge: "Low Stock" coral pill when <= 3 units
- Tap → full-screen product detail (slide from right)

**Product Detail (Buyer)**
```
┌─────────────────────────┐
│ ←                   Share│  transparent top bar over image
│                         │
│    ┌─────────────────┐  │
│    │                 │  │
│    │  Product Image  │  │  hero image — 60% of viewport
│    │  (full width)   │  │
│    │                 │  │
│    └─────────────────┘  │
│                         │
│  Black Abaya            │  product name — text-2xl semibold
│  MVR 650.00             │  price — mono, amethyst-600, text-xl
│                         │
│  Elegant black abaya... │  description — text-sm, slate-500
│                         │
│  ┌─S──┬──M──┬──L──┐    │  variant pills — horizontal
│  └────┴─────┴─────┘    │  selected = amethyst-500 fill
│   3 available           │  stock count under pills
│                         │
│   ┌─ - ─┐  2  ┌─ + ─┐  │  quantity selector
│   └─────┘     └─────┘  │
│                         │
│ ┌───────────────────────┤
│ │    Buy Now — MVR 1300 ││  sticky bottom CTA
│ └───────────────────────┤  amethyst-500 bg, white text, full-width
└─────────────────────────┘  rounded-lg, height 52px
```

**Delivery Form (Bottom Sheet)**
```
Triggered by "Buy Now" tap. Rises from bottom.

┌─────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░│  dimmed backdrop
│░░░░░░░░░░░░░░░░░░░░░░░░░│
├─────────────────────────┤
│     ══════              │  drag handle
│                         │
│  Delivery Details       │  sheet title — text-xl semibold
│                         │
│  Name                   │
│  ┌─────────────────────┐│
│  │ Ahmed Hassan        ││  text input
│  └─────────────────────┘│
│                         │
│  Phone                  │
│  ┌─────────────────────┐│
│  │ 7771234             ││  numeric input
│  └─────────────────────┘│
│                         │
│  Delivery Location      │
│  ◉ Male'  ○ Hulhumale' │  radio pills
│                         │
│  Address                │
│  ┌─────────────────────┐│
│  │                     ││  textarea
│  └─────────────────────┘│
│                         │
│  Time Preference        │
│  ┌─────────────────────┐│
│  │ After 5pm           ││  optional text
│  └─────────────────────┘│
│                         │
│ ┌───────────────────────┤
│ │  Place Order          ││  ruby-500 bg, white text
│ └───────────────────────┤
└─────────────────────────┘

Sheet specs:
- max-height: 90vh
- border-radius: radius-xl radius-xl 0 0
- shadow: shadow-sheet
- animation: slide up with ease-out, duration-sheet
- backdrop: black/40%
- drag handle: 32px wide, 4px tall, slate-300, centered
```

**Checkout (Payment)**
```
┌─────────────────────────┐
│ ←  Payment              │  top bar
├─────────────────────────┤
│                         │
│  Black Abaya — Size M   │  order summary
│  MVR 650.00             │
│                         │
│ ┌───────────────────────┤
│ │                       │
│ │   Pay with Swipe      │  PRIMARY: large amethyst button
│ │                       │  links to swipePaymentUrl
│ └───────────────────────┤
│                         │
│  or scan QR code        │  secondary: smaller QR
│  ┌─────────────────┐   │
│  │    ▓▓▓▓▓▓▓▓     │   │  QR image — 180x180px centered
│  │    ▓▓▓▓▓▓▓▓     │   │
│  │    ▓▓▓▓▓▓▓▓     │   │
│  └─────────────────┘   │
│                         │
│  ┌─ ● ─────────────┐   │  payment status
│  │ Waiting for      │   │  pulsing amethyst dot
│  │ payment...       │   │
│  └──────────────────┘   │
│                         │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │  demo controls separator
│  Demo Controls          │  slate-400 text, only visible
│  [Complete] [Expire]    │  when SWIPE_DEMO_MODE or
│  [Cancel]               │  NEXT_PUBLIC_SHOW_DEMO_CONTROLS
└─────────────────────────┘
```

**Success Page**
```
┌─────────────────────────┐
│                         │
│         ✓               │  animated checkmark — green burst
│                         │
│  Payment Confirmed!     │  text-2xl semibold
│                         │
│  Order #a8f3b2c1        │  mono, slate-500
│                         │
│ ┌───────────────────────┤
│ │ Black Abaya — Size M  │  product card
│ │ Qty: 1                │
│ │ MVR 650.00            │
│ └───────────────────────┤
│                         │
│ ┌───────────────────────┤
│ │ Delivery              │  delivery summary
│ │ Ahmed Hassan          │
│ │ Male' — H. Sunrise    │
│ │ After 5pm             │
│ └───────────────────────┤
│                         │
│ ┌───────────────────────┤
│ │  💬 WhatsApp Seller   │  WhatsApp deep link — green
│ └───────────────────────┤
│ ┌───────────────────────┤
│ │  📷 Instagram DM      │  Instagram deep link — gradient
│ └───────────────────────┤
│                         │
│ ┌───────────────────────┤
│ │  ← Back to Store      │  text button, amethyst-500
│ └───────────────────────┤
└─────────────────────────┘
```

**Seller Dashboard (Mobile)**
```
┌─────────────────────────┐
│  Island Finds MV    🔔3 │  store name + new orders badge
├─────────────────────────┤
│                         │
│  Today's Sales          │
│  MVR 2,600.00           │  large mono, amethyst-700
│                         │
│ ┌──────────┬────────────┤
│ │ Paid     │ Pending    │  2-col stat cards
│ │    4     │    1       │  large number, slate-600 label
│ ├──────────┼────────────┤
│ │ Ship     │ Low Stock  │
│ │    2     │    3       │  coral badge on low stock
│ └──────────┴────────────┘
│                         │
│  Swipe Wallet           │  amethyst gradient card
│ ┌───────────────────────┤
│ │ Available             │
│ │ MVR 15,420.50         │  white text on amethyst
│ │ Pending: MVR 650.00   │  amethyst-200 text
│ └───────────────────────┤
│                         │
│  Recent Transactions    │
│  ┌──────────────────┐   │
│  │ REF-001  +650.00 │   │  transaction rows
│  │ Fee: -19.50      │   │  net highlighted
│  └──────────────────┘   │
│                         │
├─────────────────────────┤
│ 📊  🛍️  📦  📱  ⚙️   │  bottom tab bar
└─────────────────────────┘
```

---

## Desktop Layout (>= 768px) — Web Application

### Buyer Pages (Desktop)

```
┌──────────────────────────────────────────────────────┐
│  🏝️ Island Finds MV          Search    Share    Cart │  top nav bar
├──────────────────────────────────────────────────────┤
│ All  │  Fashion  │  Gifts  │  Accessories  │  Food  │  category tabs
├──────────┬──────────┬──────────┬─────────────────────┤
│          │          │          │                     │
│ Product  │ Product  │ Product  │  Product            │  4-column grid
│ Card     │ Card     │ Card     │  Card               │  on large screens
│          │          │          │                     │  3-col on medium
├──────────┼──────────┼──────────┤                     │
│          │          │          │                     │
│ Product  │ Product  │          │                     │
│ Card     │ Card     │          │                     │
│          │          │          │                     │
└──────────┴──────────┴──────────┴─────────────────────┘
```

**Product Detail (Desktop)**
```
┌──────────────────────────────────────────────────────┐
│  ← Back to Store        Island Finds MV              │
├───────────────────────────┬──────────────────────────┤
│                           │                          │
│                           │  Black Abaya             │
│     Product Image         │  MVR 650.00              │
│     (large, left side)    │                          │
│     520 x 520px           │  Elegant black abaya...  │
│                           │                          │
│                           │  Size: [S] [M] [L]      │
│                           │  3 available             │
│                           │                          │
│                           │  Qty: [- 1 +]           │
│                           │                          │
│                           │  ┌──────────────────┐   │
│                           │  │  Buy Now          │   │
│                           │  │  MVR 650.00       │   │
│                           │  └──────────────────┘   │
│                           │                          │
│                           │  Delivery Form           │
│                           │  (inline, below Buy Now  │
│                           │   when triggered)        │
└───────────────────────────┴──────────────────────────┘

On desktop: delivery form expands INLINE below the product
info column, not as a bottom sheet.
```

**Checkout (Desktop)**
```
┌──────────────────────────────────────────────────────┐
│  ← Back                  Payment                     │
├───────────────────────────┬──────────────────────────┤
│                           │                          │
│    QR Code                │  Order Summary           │
│    ┌─────────────────┐    │                          │
│    │                 │    │  Black Abaya — Size M    │
│    │   256 x 256px   │    │  Qty: 1                  │
│    │                 │    │  MVR 650.00              │
│    └─────────────────┘    │                          │
│                           │  ─────────────────       │
│    Scan to pay with       │                          │
│    Swipe                  │  Delivery Details        │
│                           │  Ahmed Hassan            │
│    or                     │  Male' — H. Sunrise      │
│                           │                          │
│    [Pay with Swipe →]     │  ┌─ ● ────────────┐     │
│                           │  │ Waiting for     │     │
│    Demo Controls          │  │ payment...      │     │
│    [Complete][Expire]     │  └─────────────────┘     │
│    [Cancel]               │                          │
└───────────────────────────┴──────────────────────────┘

Desktop: QR is PRIMARY (left, large), link is secondary.
Two-column layout with order summary on right.
```

### Seller Pages (Desktop)

```
┌──────┬───────────────────────────────────────────────┐
│      │  Dashboard                             🔔 3   │
│  📊  ├───────────────────────────────────────────────┤
│ Dash │                                               │
│      │  ┌─────────┬─────────┬─────────┬─────────┐   │
│  🛍️  │  │ Sales   │ Paid    │ Ship    │ Low     │   │
│ Prod │  │ MVR 2.6k│   4     │   2     │ Stock 3 │   │
│      │  └─────────┴─────────┴─────────┴─────────┘   │
│  📦  │                                               │
│ Order│  ┌──────────────────────┬────────────────────┐│
│      │  │  Swipe Wallet        │ Recent Transactions ││
│  📱  │  │                      │                    ││
│Social│  │  Available           │ REF-001  +650.00   ││
│      │  │  MVR 15,420.50      │ Fee: -19.50        ││
│  ⚙️  │  │  Pending: 650.00    │ Net: 630.50        ││
│ Set  │  │                      │                    ││
│      │  └──────────────────────┴────────────────────┘│
│      │                                               │
└──────┴───────────────────────────────────────────────┘

Sidebar: 64px wide (icons only) or 240px expanded.
Icon + label vertical stack.
Active item: amethyst-500 left border + amethyst-50 background.
Content area: max-width 1200px, centered, generous padding.
```

**Seller Order List (Desktop)**
```
┌──────┬───────────────────────────────────────────────┐
│      │  Orders                                 🔔 3  │
│ side ├───────────────────────────────────────────────┤
│ bar  │  All │ Paid │ Shipped │ Delivered │ Cancelled │
│      ├──────┬────────┬───────┬────────┬──────┬──────┤
│      │  ID  │ Buyer  │ Item  │ Amount │Status│Action│
│      ├──────┼────────┼───────┼────────┼──────┼──────┤
│      │ a8f3 │ Ahmed  │ Abaya │ 650.00 │ Paid │[Ship]│
│      │ b2c1 │ Fatima │ Gift  │ 450.00 │ Paid │[Ship]│
│      │ c3d2 │ Ali    │ Case  │ 120.00 │ Sent │[Done]│
│      └──────┴────────┴───────┴────────┴──────┴──────┘
│                                                      │
└──────┴───────────────────────────────────────────────┘

Status badges: color-coded pills
  paid = amethyst-100 bg + amethyst-700 text
  shipped = blue-100 bg + blue-700 text
  delivered = slate-100 bg + slate-600 text
  cancelled = red-100 bg + red-700 text
  pending = ruby-100 bg + ruby-700 text
```

---

## Landing Page (Both Layouts)

### Mobile Landing
```
┌─────────────────────────┐
│                         │
│  Turn your DMs into     │  hero headline
│  a digital storefront   │  text-3xl, semibold
│                         │
│  Powered by             │  
│  Swipe Payments         │  amethyst-500 text
│                         │
│  [View Store →]         │  amethyst-500 pill button
│  [Seller Login]         │  text button below
│                         │
├─────────────────────────┤
│                         │
│  Before           After │  side-by-side comparison
│  ┌────────┐ ┌─────────┐│
│  │ DM     │ │ Store   ││  stacked cards
│  │ Manual │ │ Swipe   ││
│  │ Check  │ │ Auto    ││
│  │ Verify │ │ Done ✓  ││
│  └────────┘ └─────────┘│
│                         │
├─────────────────────────┤
│  3 Steps                │
│  1. Browse → 2. Pay     │  horizontal scroll
│  → 3. Confirmed         │
├─────────────────────────┤
│  No fake transfer slips │  impact stats
│  Instant reconciliation │  amethyst-500 checkmarks
│  Real-time inventory    │
├─────────────────────────┤
│  [Start Selling →]      │  final CTA — ruby-500
└─────────────────────────┘
```

### Desktop Landing
```
┌──────────────────────────────────────────────────────┐
│  🏝️ Swipe Social         [View Store] [Seller Login] │
├──────────────────────────┬───────────────────────────┤
│                          │                           │
│  Turn your Instagram     │   ┌──────────────────┐   │
│  DMs into a digital      │   │                  │   │
│  storefront              │   │  Product mockup  │   │
│                          │   │  or phone frame  │   │
│  Powered by Swipe        │   │  showing the     │   │
│  Payments                │   │  storefront      │   │
│                          │   │                  │   │
│  [View Demo Store →]     │   └──────────────────┘   │
│                          │                           │
├──────────────────────────┴───────────────────────────┤
│                                                      │
│  Before                           After              │
│  ┌────────────────────┐  ┌──────────────────────┐   │
│  │ 1. Customer DMs    │  │ 1. Browse storefront │   │
│  │ 2. Check stock     │  │ 2. Select & pay      │   │
│  │ 3. Send bank info  │  │ 3. Swipe confirms    │   │
│  │ 4. Get slip photo  │  │ 4. Order confirmed   │   │
│  │ 5. Verify manually │  │ 5. Stock auto-updated│   │
│  └────────────────────┘  └──────────────────────┘   │
│  ❌ 15 min per order       ✅ Under 2 minutes       │
│                                                      │
├──────────────────────────────────────────────────────┤
│  No fake slips  ·  Instant reconciliation  ·  Live   │
│  inventory  ·  Built for Male' & Hulhumale'          │
├──────────────────────────────────────────────────────┤
│              [Start Selling with Swipe →]             │
└──────────────────────────────────────────────────────┘
```

---

## Component Specifications

### Button Variants

```
Primary:    bg amethyst-500, text white, hover amethyst-600, radius-md, h-12 (mobile) h-10 (desktop)
Secondary:  bg white, border amethyst-500, text amethyst-500, hover amethyst-50, radius-md
Coral CTA:  bg ruby-500, text white, hover ruby-600, radius-md (for final CTAs)
Ghost:      bg transparent, text amethyst-500, hover amethyst-50, radius-md
Danger:     bg white, border red-500, text red-600, hover red-50, radius-md
Disabled:   bg slate-100, text slate-400, cursor-not-allowed
```

### Badge / Status Pill

```
Paid:       bg amethyst-100, text amethyst-700, radius-full, px-3 py-1
Shipped:    bg blue-100, text blue-700
Delivered:  bg slate-100, text slate-600
Pending:    bg ruby-100, text ruby-700
Cancelled:  bg red-100, text red-700
Low Stock:  bg ruby-100, text ruby-700, font-medium
Out:        bg slate-100, text slate-500, strikethrough
```

### Input Fields

```
Default:  border slate-200, radius-sm, h-11, px-4, text-base, font-light
Focus:    border amethyst-500, ring-2 ring-amethyst-100
Error:    border red-500, ring-2 ring-red-100
Error msg: text-sm, text red-600, mt-1
Label:    text-sm, font-medium, text slate-700, mb-1
```

### Product Card

```
Mobile:
  Container: radius-xl, overflow-hidden, bg white
  Image: square aspect-ratio, object-cover
  Info: px-3 py-3
  Name: text-sm font-medium text slate-900, line-clamp-2
  Price: text-sm font-mono text amethyst-600
  Stock badge: absolute top-2 right-2

Desktop:
  Same structure but slightly larger padding
  Image: 1:1 ratio but larger
  Hover: scale(1.02), shadow-md, transition duration-normal
```

### Bottom Sheet (Mobile Only)

```
Overlay: bg black/40%, backdrop-blur-sm
Sheet: bg white, radius-xl radius-xl 0 0, shadow-sheet
Handle: w-8 h-1 bg slate-300 radius-full, centered, mt-3
Content: px-6 pt-2 pb-safe
Animation: translateY(100%) → translateY(0), ease-out, duration-sheet
Close: tap overlay or drag down past 30% threshold
```

### Wallet Card

```
Background: linear-gradient(135deg, amethyst-600, amethyst-500)
Radius: radius-lg
Padding: space-6
Available: text white, text-2xl font-mono font-bold
Pending: text amethyst-200, text-sm
Label: text amethyst-100, text-xs font-medium uppercase tracking-wider
```

### MVR Amount Display

```
// Consistent across entire app
Font: font-mono
Format: "MVR " + number.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
Examples: "MVR 650.00", "MVR 15,420.50", "MVR 0.00"
Color: context-dependent (amethyst-600 for prices, white on dark cards, slate-900 in tables)
```

---

## Animation Specifications

### Page Transitions (Mobile)

```css
/* Push right (entering) */
@keyframes slideInRight {
  from { transform: translateX(100%); opacity: 0.8; }
  to { transform: translateX(0); opacity: 1; }
}

/* Push left (exiting) */
@keyframes slideOutLeft {
  from { transform: translateX(0); opacity: 1; }
  to { transform: translateX(-30%); opacity: 0.6; }
}

/* Duration: duration-normal, easing: ease-out */
```

### Bottom Sheet

```css
/* Rise */
@keyframes sheetRise {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}
/* Duration: duration-sheet, easing: ease-spring */

/* Backdrop */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

### Payment Status

```css
/* Pending pulse */
@keyframes pendingPulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.3); opacity: 0.5; }
}
/* Applied to amethyst dot indicator, 2s infinite */

/* Success checkmark burst */
@keyframes checkBurst {
  0% { transform: scale(0); opacity: 0; }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); opacity: 1; }
}
/* green circle with check, ease-spring, 500ms */
```

### Card Interactions

```css
/* Mobile: press feedback */
.product-card:active {
  transform: scale(0.97);
  transition: transform var(--duration-fast) var(--ease-smooth);
}

/* Desktop: hover */
.product-card:hover {
  transform: scale(1.02);
  box-shadow: var(--shadow-md);
  transition: all var(--duration-normal) var(--ease-smooth);
}
```

---

## Tailwind Configuration

```ts
// tailwind.config.ts
import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        amethyst: {
          50: '#faf5ff', 100: '#f3e8ff', 200: '#e9d5ff',
          300: '#d8b4fe', 400: '#c084fc', 500: '#a855f7',
          600: '#9333ea', 700: '#7e22ce', 800: '#6b21a8',
          900: '#581c87',
        },
        ruby: {
          50: '#fff1f2', 100: '#ffe4e6', 200: '#fecdd3',
          300: '#fda4af', 400: '#fb7185', 500: '#f43f5e',
          600: '#e11d48', 700: '#be123c',
        },
        slate: {
          50: '#f8fafc', 100: '#f1f5f9', 200: '#e2e8f0',
          300: '#cbd5e1', 400: '#94a3b8', 500: '#64748b',
          600: '#475569', 700: '#334155', 800: '#1e293b',
          900: '#0f172a',
        },
      },
      fontFamily: {
        display: ['Inter', 'sans-serif'],
        body: ['DM Sans', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      borderRadius: {
        sm: '0.375rem', md: '0.75rem', lg: '1rem',
        xl: '1.5rem',
      },
      boxShadow: {
        sm: '0 1px 2px rgba(15,23,42,0.04)',
        md: '0 4px 12px rgba(15,23,42,0.07)',
        lg: '0 8px 24px rgba(15,23,42,0.09)',
        xl: '0 16px 48px rgba(15,23,42,0.12)',
        sheet: '0 -4px 24px rgba(15,23,42,0.14)',
      },
      colors: {
        // ... amethyst, ruby, slate as above, plus:
        blue: {
          100: '#dbeafe', 700: '#1d4ed8',  // for shipped badge
        },
      },
      transitionDuration: {
        fast: '150ms',
        normal: '250ms',
        slow: '400ms',
        sheet: '350ms',
      },
      transitionTimingFunction: {
        'out-expo': 'cubic-bezier(0.16, 1, 0.3, 1)',
        'spring': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
      zIndex: {
        60: '60',  // bottom sheet overlay
        70: '70',  // bottom sheet content
      },
    },
  },
  plugins: [],
}

export default config
```

---

## File Structure Additions

```
components/
├── shells/
│   ├── mobile-shell.tsx      # Bottom tab bar, safe areas, page transitions
│   ├── desktop-shell.tsx     # Top nav (buyer) or sidebar (seller)
│   └── layout-router.tsx     # JS-based viewport detection, shell selection
│   └── splash-loader.tsx     # Brief loading state during SSR → hydration
├── mobile/
│   ├── bottom-tab-bar.tsx    # Seller bottom navigation
│   ├── top-bar.tsx           # Back arrow + title + action
│   ├── bottom-sheet.tsx      # Reusable sheet with drag-to-dismiss
│   └── page-transition.tsx   # Slide-in/out wrapper
├── desktop/
│   ├── seller-sidebar.tsx    # Icon + label sidebar
│   ├── buyer-top-nav.tsx     # Store header + nav
│   └── content-area.tsx      # Max-width centered content wrapper
```

---

## Implementation Notes

1. **Font loading**: Add `Inter` (weights 300-800), `DM Sans` (weights 300-700), and `JetBrains Mono` (weight 400) via `next/font/google` in root layout.

2. **Safe areas**: Mobile shell uses `env(safe-area-inset-top)` and `env(safe-area-inset-bottom)` for notch/home-indicator devices.

3. **No CSS media queries for layout switching**: The `useIsMobile()` hook runs once on mount. The root `LayoutRouter` renders either `MobileShell` or `DesktopShell`. Components within each shell can use Tailwind's responsive utilities for MINOR adjustments (text size, padding), but the structural layout is determined by the shell, not by breakpoints.

4. **Image optimization**: Use `next/image` with `sizes` prop set per layout. Mobile: `(max-width: 768px) 50vw`. Desktop: `25vw`.

5. **Seller tab bar z-index**: `z-50` to stay above page content. Shadow-sm on top border.

6. **Bottom sheet z-index**: `z-60` (overlay) + `z-70` (sheet) to float above tab bar.
