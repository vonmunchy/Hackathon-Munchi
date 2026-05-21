---
title: "Dual-Layout Shell: JS Viewport Detection for Native-Feel Mobile + Web Desktop"
date: 2026-05-20
last_updated: 2026-05-21
category: architecture-patterns
module: swipe-social-storefront
problem_type: architecture_pattern
component: tooling
severity: medium
applies_when:
  - "Mobile must feel like a native app (bottom nav, bottom sheets, slide transitions) while desktop is a traditional web-app shell"
  - "CSS responsive breakpoints would force layout compromises that degrade the native-app feel on mobile"
  - "A single Next.js page tree must serve two structurally different layout shells without code duplication per page"
  - "Mobile and desktop shells differ in DOM structure, navigation paradigm, and interaction model — not just styling"
tags:
  - nextjs
  - layout
  - mobile
  - desktop
  - viewport-detection
  - dual-shell
  - native-app-feel
  - social-commerce
---

# Dual-Layout Shell: JS Viewport Detection for Native-Feel Mobile + Web Desktop

## Context

Building Swipe Social Storefront — a social commerce platform for Maldives Instagram/Facebook sellers — required mobile to feel genuinely native (bottom navigation, bottom sheet modals, slide-in page transitions, touch-optimized interactions) while desktop served as a professional seller command center (sidebar navigation, multi-column grids, inline forms).

A single responsive layout using CSS media queries cannot deliver this. Responsive CSS only restyls existing DOM — it cannot restructure navigation paradigms, replace a sidebar with a bottom tab bar, swap inline forms for rising bottom sheets, or add slide-in page transitions. The mobile experience with responsive CSS feels like a shrunken desktop site, not a real app. (auto memory [claude]: user explicitly rejected responsive design for this exact reason.)

The target audience reinforced this requirement: Maldivian social sellers and their buyers are mobile-first users who expect app-like experiences from platforms they use daily (Instagram, WhatsApp). The seller dashboard needed to be equally functional on desktop where sellers manage inventory, orders, and financials.

## Guidance

**The pattern: JS-based viewport detection at mount time routes to one of two completely separate layout shells.**

### Step 1 — useIsMobile() hook

```tsx
// lib/use-device.ts
'use client'
import { useState, useEffect } from 'react'

export function useIsMobile(): boolean | null {
  const [isMobile, setIsMobile] = useState<boolean | null>(null)

  useEffect(() => {
    const check = () => window.innerWidth < 768
    setIsMobile(check())
    // Do NOT add resize listener — layout is fixed for the session
  }, [])

  return isMobile  // null = SSR/loading, true = mobile, false = desktop
}
```

No resize listener is added. The layout decision is made once at mount and stays fixed for the session. If the user rotates a tablet, they keep whichever shell loaded. This prevents the shell from tearing mid-session.

### Step 2 — LayoutRouter

```tsx
// components/shells/layout-router.tsx
'use client'
export function LayoutRouter({ children }) {
  const isMobile = useIsMobile()
  if (isMobile === null) return <SplashLoader />  // SSR → hydration gap
  if (isMobile) return <MobileShell>{children}</MobileShell>
  return <DesktopShell>{children}</DesktopShell>
}
```

The SplashLoader handles the SSR gap where `window` is not yet available.

### Step 3 — Root layout wraps everything in LayoutRouter

```tsx
// app/layout.tsx
export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <LayoutRouter>{children}</LayoutRouter>
      </body>
    </html>
  )
}
```

### MobileShell structure

- Full-screen content area with safe-area-inset padding
- **Buyer pages**: BuyerShell with context-aware TopBar (title + back button from `getBuyerTopBarConfig(pathname)`) + BuyerBottomBar (Home, Marketplace, Exchange). Safe area handled via `padding-top: max(12px, env(safe-area-inset-top))`.
- **Seller pages**: TopBar + bottom tab bar (56px + safe-area-inset-bottom) with Dashboard, Products, Pay Links, Exchange, Orders, Settings
- Page transitions: slide in from right (ease-out, 250ms)
- Bottom sheets for modals: max-height 90vh, spring animation (350ms), drag-to-dismiss at 30% threshold

### DesktopShell structure

- **Buyer pages**: BuyerDesktopNav — sticky top bar with SwiftStore logo, Marketplace/Exchange links, "Become a Seller" CTA
- **Seller pages**: SellerSidebar — fixed 260px sidebar with 7 nav links (Dashboard, Products, Payment Links, Exchange, Orders, Social, Settings)
- Content area: max-width varies by page (4xl-5xl), centered, p-4 md:p-6 padding
- **Backstage (admin)**: no shell — renders raw with its own dark-themed layout

### File structure

```
components/
├── shells/
│   ├── layout-router.tsx
│   ├── mobile-shell.tsx
│   ├── desktop-shell.tsx
│   └── splash-loader.tsx
├── mobile/
│   ├── bottom-tab-bar.tsx
│   ├── top-bar.tsx
│   ├── bottom-sheet.tsx
│   └── page-transition.tsx
├── desktop/
│   ├── seller-sidebar.tsx
│   ├── buyer-top-nav.tsx
│   └── content-area.tsx
lib/
└── use-device.ts
```

### Critical constraint

No CSS media queries are used for shell selection. Tailwind responsive utilities (`md:`, `lg:`) are permitted for minor adjustments within each shell (text size, padding), but the structural layout — which shell renders — is determined entirely by the JS hook.

## Why This Matters

Responsive CSS breakpoints only change visual styling. They cannot give you:

- A persistent bottom tab bar that stays fixed while page content scrolls above it
- Bottom sheet modals that rise from below with spring physics and drag-to-dismiss
- Full-screen page transitions that slide in from the right like a native navigation stack
- Safe area insets for notch and home-indicator devices
- A layout that feels genuinely native rather than a desktop site that stacks on small screens

These are structural differences in how the DOM is organized, how scroll is managed, how navigation works, and how modals appear. CSS alone cannot restructure the DOM based on viewport width.

The dual-shell approach means mobile users get a component tree built for mobile and desktop users get a component tree built for desktop. The same page content (the `children`) flows into whichever shell is appropriate.

## When to Apply

- Mobile and desktop users have fundamentally different interaction patterns, not just different amounts of screen space
- Mobile users expect native-app behaviors: bottom navigation, bottom sheets, swipe gestures, slide transitions
- Desktop users expect web-application behaviors: sidebars, data tables, multi-column layouts, hover states, inline forms
- The product is mobile-first but desktop matters enough to build properly
- Product categories: social commerce, consumer marketplaces, field tools, food ordering, apps with distinct buyer/seller roles

**Do NOT apply** when both platforms need essentially the same interaction pattern with only density differences (content sites, blogs, simple form tools). Responsive CSS is simpler and sufficient there.

## Examples

**Before (responsive):** Single layout with `grid-cols-1 md:grid-cols-2 lg:grid-cols-4`. Sidebar collapses into hamburger on mobile. Forms appear inline on both. The mobile experience works but feels like a desktop site that was squished.

**After (dual-shell):**
- **Mobile buyer**: clean full-screen storefront with back arrow, 2-column product grid, product detail slides in from right, bottom sheet rises for delivery form, prominent "Pay with Swipe" button
- **Mobile seller**: bottom tab bar (Dashboard/Products/Orders/Social/Settings), pages transition smoothly, notification badge on Orders tab
- **Desktop buyer**: top nav with store name, category tabs, 4-column product grid, two-column product detail (large image left, details right), inline delivery form
- **Desktop seller**: collapsible sidebar, multi-column dashboard with stat cards, data tables for order management

Same Convex queries. Same API routes. Same product data. Two completely separate structural shells that each feel native to their platform.

## Related

- **Sibling architecture pattern**: `docs/solutions/architecture-patterns/swipe-nextjs-convex-split-responsibility-2026-05-20.md` — covers the payment integration network topology (Next.js as network bridge to Swipe mock + Convex for persistence). Section 4 of that doc advises different QR vs link display by viewport, which is a display-level concern that this dual-shell pattern implements structurally.
- **Design system**: `docs/design-system.md` — complete token definitions, wireframes, and component specs for both shells
- **Implementation plan U1.5**: `docs/plans/2026-05-20-001-feat-swipe-social-storefront-plan.md` — the layout shell implementation unit
