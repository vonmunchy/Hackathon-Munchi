---
title: "BML x Swipe Brand Fusion Design System"
date: 2026-05-21
last_updated: 2026-05-21
category: design-patterns
module: swipe-social-storefront
problem_type: design_pattern
component: frontend_stimulus
severity: medium
applies_when:
  - building a product under an established parent brand that needs to reference both identities
  - migrating design tokens to match real brand guidelines from research
  - targeting a market where institutional trust and fintech credibility are both critical
  - the user specifies a premium feel without ornamental elements
  - the application has distinct mobile and desktop shells needing consistent branding
tags:
  - design-system
  - brand-identity
  - color-tokens
  - typography
  - navigation
  - tailwind-css
  - next-js
  - premium-ui
---

# BML x Swipe Brand Fusion Design System

## Context

The Swipe Social Storefront needed a cohesive design system that reflects its dual brand heritage: Bank of Maldives (BML) as the institutional parent (crimson red #CC0D0D, corporate trust) and Swipe by BML as the modern fintech product (violet purple #702FC8, digital innovation). The existing color tokens (ocean/teal, coral/orange, sand/warm-gray) did not align with either brand and gave the app a generic feel. The user explicitly required a "lux and premium, clean, no gold" aesthetic. (auto memory [claude]) The product vision is seller-centric social commerce for Maldives social sellers, and the dual-layout architecture (distinct mobile app-like shell and desktop sidebar shell) was a firm constraint.

## Guidance

**Derive design system tokens directly from the real brand assets of the product and its parent company, then map them to semantic roles.**

### Color Token Strategy

Use evocative names (gemstone/material metaphor) rather than generic or raw brand names:

- **Amethyst** (primary, from Swipe's violet): 50 #faf5ff through 900 #581c87
- **Ruby** (accent, from BML's crimson): 50 #fff1f2 through 700 #be123c
- **Slate** (neutral, cool blue-gray): 50 #f8fafc through 900 #0f172a

Each palette needs a full 50-900 scale defined as CSS custom properties in a Tailwind v4 `@theme inline` block.

### Semantic Color Roles

Beyond the core palette, specific semantic meanings are assigned:

- **Success green** (`#059669`) — Buy actions, positive status, confirmations. Used for exchange Buy tab and "Buy USDT" buttons.
- **Ruby red** — Sell actions, destructive actions (reset, delete), error states. Used for exchange Sell tab and admin Reset button.
- **Amethyst** — Primary brand, active navigation, CTAs, links. The landing page uses amethyst as the sole accent on a near-monochrome base.
- **Slate** — All neutral surfaces, borders (hairline at low alpha: `rgba(15,23,42,0.06)`), text hierarchy.

### Exchange-Specific Tokens

The P2P exchange uses a Binance-inspired Buy/Sell color paradigm mapped to the design system:

- Buy tab active: `bg-success text-white` (green)
- Sell tab active: `bg-ruby-500 text-white` (red)
- Inactive tab: `text-slate-500 bg-transparent`
- Order book table: `font-mono` for all prices/amounts, `truncate` for advertiser names on mobile

### Shadow Treatment

Shift shadows from warm to cool by using the slate-900 base color:

```css
/* Before - generic warm */
--shadow-md: 0 4px 12px rgba(0, 0, 0, 0.06);

/* After - cool, purple-tinted depth */
--shadow-md: 0 4px 12px rgba(15, 23, 42, 0.07);
```

### Typography Pairing

Inter (display, weights 300-800) + DM Sans (body, weights 300-700) with `letter-spacing: -0.011em` on body text. Both are geometric sans-serifs common in premium fintech products.

### Mobile Shell

- Top bar: `backdrop-blur-xl`, `bg-white/80`, `border-b border-slate-100` for frosted glass
- Bottom tab bar: 60px height, HeroIcons (filled when active, outlined when inactive), max 5 tabs
- Background: `bg-slate-50` on main content area

### Desktop Shell

- Sidebar: 260px wide, branded "S" mark (amethyst-600 rounded-lg), 1.5px stroke icons, "Powered by Swipe" footer
- Active nav state: `bg-amethyst-50 text-amethyst-700` (subtle, not loud)
- Buyer top nav: same frosted glass treatment as mobile
- Background: `bg-slate-50` on content area

### Landing Page Refinements

- Uppercase "POWERED BY SWIPE" label with `tracking-[0.15em]`
- Heading `tracking-[-0.03em]` for tighter premium feel
- Buttons: `rounded-xl` (not `rounded-full`) for modern refinement
- Final CTA: dark `bg-slate-900` instead of brand color for understated elegance

### Bulk Migration Approach

Use sed find-and-replace for token names across all component files:
```bash
sed -i 's/ocean-/amethyst-/g; s/coral-/ruby-/g; s/sand-/slate-/g' **/*.tsx
```
Then manually refine key shell and page components where the replacement needs design judgment beyond color swapping.

## Why This Matters

A fintech product operating under a bank's umbrella must visually communicate both trust (inherited from the parent institution) and modernity (the product's own identity). Using off-brand colors creates visual dissonance and undermines credibility with users who know the BML/Swipe brands.

The two-brand architecture -- red for heritage trust, purple for digital innovation -- is a deliberate strategic choice by BML. The design system honors both. Cool neutrals and restrained typography reinforce "premium without flash" positioning.

The bulk sed approach is safe when the old token names (ocean, coral, sand) are unique to the design system and don't collide with other identifiers. This let us update 32 files in seconds, then focus manual effort on the 6-8 components that needed design refinement.

## When to Apply

- Building a product under an established parent brand that needs to reference both identities
- Migrating an existing design system's color tokens to match real brand guidelines
- Targeting a market where institutional trust and modern fintech credibility are both critical
- The user specifies a "luxury/premium" feel without ornamental elements (no gold, no excessive gradients)
- The application has distinct mobile (app-like) and desktop (sidebar) shells needing consistent but layout-appropriate branding

## Examples

**Color tokens (globals.css):**

Before:
```css
--color-ocean-500: #0d9488;   /* teal primary */
--color-coral-500: #f97316;   /* orange accent */
--color-sand-200: #e7e5e4;    /* warm gray neutral */
```

After:
```css
--color-amethyst-500: #a855f7;  /* Swipe violet primary */
--color-ruby-500: #f43f5e;      /* BML crimson accent */
--color-slate-200: #e2e8f0;     /* cool blue-gray neutral */
```

**Typography (layout.tsx):**

Before:
```tsx
import { Plus_Jakarta_Sans } from "next/font/google";
const font = Plus_Jakarta_Sans({ subsets: ["latin"], weight: ["400", "600", "700"] });
```

After:
```tsx
import { Inter, DM_Sans } from "next/font/google";
const inter = Inter({ variable: "--font-inter", subsets: ["latin"], weight: ["300","400","500","600","700","800"] });
const dmSans = DM_Sans({ variable: "--font-dm-sans", subsets: ["latin"], weight: ["300","400","500","600","700"] });
```

**Mobile bottom tab bar:**

Before: stroke-only SVG icons, 56px height, `text-ocean-500` active
After: HeroIcons filled/outlined, 60px height, `text-amethyst-600` active, `backdrop-blur-xl`

**Desktop sidebar active state:**

Before: `border-l-2 border-ocean-500 bg-ocean-50 text-ocean-500`
After: `bg-amethyst-50 text-amethyst-700` (no left border, softer)

**Landing page CTA:**

Before: `bg-coral-500 rounded-full px-8 py-4 text-lg font-bold`
After: `bg-slate-900 rounded-xl px-10 py-5 text-lg font-semibold`

## Related

- `docs/design-system.md` -- The original design system spec document. **NEEDS REFRESH**: still references ocean/coral/sand palette and Plus Jakarta Sans typography which are now superseded by the amethyst/ruby/slate palette and Inter/DM Sans.
- `docs/solutions/architecture-patterns/dual-layout-shell-js-viewport-detection-2026-05-20.md` -- Describes the dual mobile/desktop shell architecture that this design system operates within. Layout structure unchanged; only visual tokens affected.
