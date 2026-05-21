'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { useState } from 'react'

function FadeIn({
  children,
  className = '',
  delay = 0,
}: {
  children: React.ReactNode
  className?: string
  delay?: number
}) {
  return (
    <motion.div
      initial={false}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: '-48px' }}
      transition={{ duration: 0.4, delay, ease: [0.16, 1, 0.3, 1] }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

function MenuIcon({ open }: { open: boolean }) {
  return (
    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
      {open ? (
        <path
          fillRule="evenodd"
          d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
          clipRule="evenodd"
        />
      ) : (
        <path
          fillRule="evenodd"
          d="M3 5a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 15a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"
          clipRule="evenodd"
        />
      )}
    </svg>
  )
}

/* ── Feature card icons ── */
function StorefrontIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  )
}

function PaymentIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="3" height="3" />
      <line x1="21" y1="14" x2="21" y2="14.01" />
      <line x1="21" y1="21" x2="21" y2="21.01" />
    </svg>
  )
}

function ExchangeIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="17 1 21 5 17 9" />
      <path d="M3 11V9a4 4 0 014-4h14" />
      <polyline points="7 23 3 19 7 15" />
      <path d="M21 13v2a4 4 0 01-4 4H3" />
    </svg>
  )
}

function OrdersIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M9 11l3 3L22 4" />
      <path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" />
    </svg>
  )
}

function SocialIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M18 2h-3a5 5 0 00-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 011-1h3z" />
    </svg>
  )
}

function ShieldIcon() {
  return (
    <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <polyline points="9 12 11 14 15 10" />
    </svg>
  )
}

/* ── Page ── */

export default function Home() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const capabilities = [
    {
      icon: <StorefrontIcon />,
      title: 'Instant Storefronts',
      desc: 'Products, variants, stock tracking, and a shareable link ready for Instagram bios.',
      color: 'bg-amethyst-100 text-amethyst-700',
    },
    {
      icon: <PaymentIcon />,
      title: 'Swipe QR Checkout',
      desc: 'Customers scan, pay in MVR, and get instant confirmation. No bank transfers.',
      color: 'bg-emerald-100 text-emerald-700',
    },
    {
      icon: <ExchangeIcon />,
      title: 'P2P USDT Exchange',
      desc: 'Built-in order book for USDT/MVR trading with escrow-backed reservations.',
      color: 'bg-blue-100 text-blue-700',
    },
    {
      icon: <OrdersIcon />,
      title: 'Order Management',
      desc: 'Track orders from payment to delivery. Customers get live status updates.',
      color: 'bg-amber-100 text-amber-700',
    },
    {
      icon: <SocialIcon />,
      title: 'Social Publishing',
      desc: 'Generate captions and publish products directly to Facebook and Instagram.',
      color: 'bg-pink-100 text-pink-700',
    },
    {
      icon: <ShieldIcon />,
      title: 'Multi-Tenant Security',
      desc: 'Each seller gets isolated data, session auth, and their own Swipe credentials.',
      color: 'bg-slate-100 text-slate-700',
    },
  ]

  const metrics = [
    { value: '< 2 min', label: 'Store setup' },
    { value: 'Instant', label: 'Payment confirmation' },
    { value: '6', label: 'Core modules' },
    { value: '0%', label: 'Platform fee' },
  ]

  return (
    <div className="min-h-screen overflow-x-hidden bg-white font-body text-slate-900">
      {/* Nav */}
      <nav className="sticky top-0 z-50 border-b border-slate-200/70 bg-white/85 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-5 sm:px-8">
          <Link href="/" className="flex items-center gap-3" aria-label="SwiftStore home">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-amethyst-600 font-display text-lg font-bold text-white">
              S
            </span>
            <span className="font-display text-lg font-bold tracking-tight text-slate-900">SwiftStore</span>
          </Link>

          <div className="hidden items-center gap-8 md:flex">
            <Link href="/marketplace" className="text-sm font-medium text-slate-500 transition-colors hover:text-slate-900">
              Marketplace
            </Link>
            <Link href="/exchange" className="text-sm font-medium text-slate-500 transition-colors hover:text-slate-900">
              Exchange
            </Link>
            <Link
              href="/seller/login"
              className="rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-slate-800"
            >
              Start Selling
            </Link>
          </div>

          <button
            onClick={() => setMobileMenuOpen((open) => !open)}
            className="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-50 hover:text-slate-900 md:hidden"
            aria-label="Menu"
            aria-expanded={mobileMenuOpen}
          >
            <MenuIcon open={mobileMenuOpen} />
          </button>
        </div>

        {mobileMenuOpen && (
          <div className="space-y-2 border-t border-slate-200/70 bg-white px-5 py-4 md:hidden">
            {[
              ['Marketplace', '/marketplace'],
              ['Exchange', '/exchange'],
              ['Start Selling', '/seller/login'],
            ].map(([label, href]) => (
              <Link
                key={href}
                href={href}
                className="block rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setMobileMenuOpen(false)}
              >
                {label}
              </Link>
            ))}
          </div>
        )}
      </nav>

      <main>
        {/* ── Hero ── */}
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-b from-amethyst-50/40 via-white to-white" aria-hidden="true" />
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(circle, #6b21a8 1px, transparent 1px)', backgroundSize: '32px 32px' }} aria-hidden="true" />

          <div className="relative mx-auto max-w-5xl px-5 pb-20 pt-24 text-center sm:px-8 sm:pb-28 sm:pt-32 lg:pb-32 lg:pt-40">
            <FadeIn>
              <div className="mb-8 inline-flex items-center gap-2.5 rounded-full border border-slate-200/80 bg-white px-4 py-2 shadow-sm">
                <span className="relative flex h-2 w-2">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75" />
                  <span className="relative inline-flex h-2 w-2 rounded-full bg-success" />
                </span>
                <span className="text-xs font-semibold text-slate-600">Powered by Swipe</span>
              </div>
            </FadeIn>

            <FadeIn delay={0.05}>
              <h1 className="mx-auto w-full max-w-[21rem] font-display text-[38px] font-semibold leading-[1] tracking-[-0.035em] text-slate-950 sm:max-w-4xl sm:text-[64px] sm:leading-[0.96] lg:text-[80px]">
                <span className="sm:hidden">
                  Social commerce
                  <br />
                  for the
                  <br />
                  Maldives.
                </span>
                <span className="hidden sm:inline">Social commerce infrastructure for the Maldives.</span>
              </h1>
            </FadeIn>

            <FadeIn delay={0.1}>
              <p className="mx-auto mt-7 w-full max-w-[21rem] text-base leading-7 text-slate-500 sm:max-w-2xl sm:text-lg sm:leading-8">
                Storefronts, Swipe payments, order management, and P2P USDT exchange — everything Maldivian social sellers need, in one platform.
              </p>
            </FadeIn>

            <FadeIn delay={0.15}>
              <div className="mx-auto mt-10 flex w-full max-w-[280px] flex-col items-center justify-center gap-3 sm:w-auto sm:max-w-none sm:flex-row">
                <Link
                  href="/seller/login"
                  className="inline-flex w-full items-center justify-center rounded-xl bg-amethyst-600 px-8 py-4 text-sm font-semibold text-white shadow-lg shadow-amethyst-200 transition-all hover:bg-amethyst-700 hover:shadow-amethyst-300 sm:w-auto"
                >
                  Start Selling
                </Link>
                <Link
                  href="/marketplace"
                  className="inline-flex w-full items-center justify-center rounded-xl border border-slate-200 bg-white px-8 py-4 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-50 sm:w-auto"
                >
                  Browse Marketplace
                </Link>
                <Link
                  href="/exchange"
                  className="inline-flex w-full items-center justify-center rounded-xl border border-slate-200 bg-white px-8 py-4 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-50 sm:w-auto"
                >
                  P2P Exchange
                </Link>
              </div>
            </FadeIn>
          </div>
        </section>

        {/* ── Metrics ── */}
        <section className="border-y border-slate-200/70 bg-slate-50">
          <div className="mx-auto max-w-6xl px-5 py-12 sm:px-8 sm:py-14">
            <div className="grid grid-cols-2 gap-8 text-center sm:grid-cols-4">
              {metrics.map((stat) => (
                <FadeIn key={stat.label}>
                  <p className="font-display text-2xl font-bold tracking-tight text-slate-900 sm:text-4xl">{stat.value}</p>
                  <p className="mt-1 text-xs font-medium text-slate-400 sm:text-sm">{stat.label}</p>
                </FadeIn>
              ))}
            </div>
          </div>
        </section>

        {/* ── Capabilities Grid ── */}
        <section className="bg-white">
          <div className="mx-auto max-w-6xl px-5 py-20 sm:px-8 lg:py-28">
            <FadeIn className="text-center">
              <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">Platform Capabilities</p>
              <h2 className="mx-auto mt-4 max-w-3xl font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
                Everything a social seller needs. Nothing they don&apos;t.
              </h2>
            </FadeIn>

            <div className="mt-14 grid gap-5 sm:mt-16 sm:grid-cols-2 lg:grid-cols-3">
              {capabilities.map((cap, i) => (
                <FadeIn key={cap.title} delay={i * 0.05}>
                  <div className="group flex h-full flex-col rounded-2xl border border-slate-200/80 bg-white p-7 transition-all hover:border-slate-300 hover:shadow-md sm:p-8">
                    <div className={`flex h-11 w-11 items-center justify-center rounded-xl ${cap.color}`}>
                      {cap.icon}
                    </div>
                    <h3 className="mt-5 font-display text-lg font-semibold text-slate-900">{cap.title}</h3>
                    <p className="mt-2 text-sm leading-6 text-slate-500">{cap.desc}</p>
                  </div>
                </FadeIn>
              ))}
            </div>
          </div>
        </section>

        {/* ── How It Works ── */}
        <section className="border-y border-slate-200/70 bg-slate-50">
          <div className="mx-auto max-w-6xl px-5 py-20 sm:px-8 lg:py-28">
            <FadeIn className="text-center">
              <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">How it works</p>
              <h2 className="mx-auto mt-4 max-w-2xl font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
                Three steps from post to payment.
              </h2>
            </FadeIn>

            <div className="mx-auto mt-14 grid max-w-4xl gap-0 sm:mt-16 md:grid-cols-3">
              {[
                { num: '01', title: 'Create your store', desc: 'Add products, set MVR prices, connect Swipe credentials. Get a shareable link in under 2 minutes.' },
                { num: '02', title: 'Share & sell', desc: 'Post your store link on Instagram or Facebook. Customers browse, pick variants, and check out instantly.' },
                { num: '03', title: 'Get paid via Swipe', desc: 'Swipe confirms payment via QR. Orders update live. Ship the product and track everything from your dashboard.' },
              ].map((step, i) => (
                <FadeIn key={step.num} delay={i * 0.08}>
                  <div className="relative flex flex-col items-center px-6 py-8 text-center md:px-8">
                    {i < 2 && (
                      <div className="absolute bottom-0 left-1/2 h-8 w-px bg-slate-200 md:bottom-auto md:left-auto md:right-0 md:top-1/2 md:h-px md:w-8 md:-translate-y-1/2" />
                    )}
                    <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-amethyst-600 font-display text-xl font-bold text-white">
                      {step.num}
                    </span>
                    <h3 className="mt-5 font-display text-lg font-semibold text-slate-900">{step.title}</h3>
                    <p className="mt-2 text-sm leading-6 text-slate-500">{step.desc}</p>
                  </div>
                </FadeIn>
              ))}
            </div>
          </div>
        </section>

        {/* ── Tech Stack ── */}
        <section className="bg-white">
          <div className="mx-auto max-w-6xl px-5 py-20 sm:px-8 lg:py-28">
            <FadeIn className="text-center">
              <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">Built with</p>
              <h2 className="mx-auto mt-4 max-w-2xl font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
                Modern stack, production-grade.
              </h2>
            </FadeIn>

            <div className="mx-auto mt-14 grid max-w-3xl grid-cols-2 gap-4 sm:mt-16 sm:grid-cols-3 md:grid-cols-6">
              {[
                { name: 'Next.js 16', sub: 'Frontend' },
                { name: 'Convex', sub: 'Backend' },
                { name: 'Tailwind', sub: 'Styling' },
                { name: 'Swipe API', sub: 'Payments' },
                { name: 'Meta API', sub: 'Social' },
                { name: 'Vercel', sub: 'Deploy' },
              ].map((tech) => (
                <FadeIn key={tech.name}>
                  <div className="flex flex-col items-center rounded-xl border border-slate-100 bg-slate-50/50 px-4 py-5 text-center">
                    <p className="font-display text-sm font-bold text-slate-900">{tech.name}</p>
                    <p className="mt-0.5 text-xs text-slate-400">{tech.sub}</p>
                  </div>
                </FadeIn>
              ))}
            </div>
          </div>
        </section>

        {/* ── CTA ── */}
        <section className="bg-slate-900 px-5 py-20 text-center sm:px-8 lg:py-28">
          <FadeIn className="mx-auto max-w-3xl">
            <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-300">Ready to explore</p>
            <h2 className="mt-4 font-display text-3xl font-semibold tracking-[-0.03em] text-white sm:text-5xl">
              See the full platform in action.
            </h2>
            <p className="mx-auto mt-5 max-w-xl text-base leading-7 text-slate-400">
              Browse the marketplace, try the seller dashboard, or check out the P2P exchange. Everything works end-to-end in demo mode.
            </p>
            <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Link
                href="/seller/login"
                className="inline-flex w-full items-center justify-center rounded-xl bg-white px-8 py-4 text-sm font-semibold text-slate-900 transition-colors hover:bg-slate-100 sm:w-auto"
              >
                Open Seller Dashboard
              </Link>
              <Link
                href="/marketplace"
                className="inline-flex w-full items-center justify-center rounded-xl border border-slate-600 px-8 py-4 text-sm font-semibold text-slate-300 transition-colors hover:border-slate-500 hover:text-white sm:w-auto"
              >
                Browse Marketplace
              </Link>
              <Link
                href="/exchange"
                className="inline-flex w-full items-center justify-center rounded-xl border border-slate-600 px-8 py-4 text-sm font-semibold text-slate-300 transition-colors hover:border-slate-500 hover:text-white sm:w-auto"
              >
                P2P Exchange
              </Link>
            </div>
          </FadeIn>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200/70 bg-white">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-5 py-8 text-center sm:flex-row sm:items-center sm:justify-between sm:px-8 sm:text-left">
          <span className="font-display text-sm font-bold tracking-tight text-slate-900">SwiftStore</span>
          <div className="flex items-center justify-center gap-6">
            <Link href="/marketplace" className="text-xs text-slate-400 transition-colors hover:text-slate-600">
              Marketplace
            </Link>
            <Link href="/exchange" className="text-xs text-slate-400 transition-colors hover:text-slate-600">
              Exchange
            </Link>
            <Link href="/seller/login" className="text-xs text-slate-400 transition-colors hover:text-slate-600">
              Sellers
            </Link>
          </div>
          <p className="text-xs text-slate-400">&copy; 2025 SwiftStore &middot; Powered by Swipe</p>
        </div>
      </footer>
    </div>
  )
}
