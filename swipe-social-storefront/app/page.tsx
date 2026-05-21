'use client'

import Image from 'next/image'
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

/* ── Icons for How It Works ── */
function StoreIcon() {
  return (
    <svg className="h-7 w-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  )
}

function QrIcon() {
  return (
    <svg className="h-7 w-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="3" height="3" />
      <line x1="21" y1="14" x2="21" y2="14.01" />
      <line x1="21" y1="21" x2="21" y2="21.01" />
      <line x1="17" y1="21" x2="17" y2="21.01" />
      <line x1="14" y1="21" x2="14" y2="21.01" />
      <line x1="21" y1="17" x2="21" y2="17.01" />
    </svg>
  )
}

function ExchangeIcon() {
  return (
    <svg className="h-7 w-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="17 1 21 5 17 9" />
      <path d="M3 11V9a4 4 0 014-4h14" />
      <polyline points="7 23 3 19 7 15" />
      <path d="M21 13v2a4 4 0 01-4 4H3" />
    </svg>
  )
}

/* ── Sections ── */

function ProductStrip() {
  const products = [
    ['/demo-products/abaya.png', 'Abaya storefront product'],
    ['/demo-products/dates-box.png', 'Dates gift box product'],
    ['/demo-products/bracelet.png', 'Bracelet storefront product'],
    ['/demo-products/phone-case.png', 'Phone case storefront product'],
  ]

  return (
    <FadeIn delay={0.2} className="mx-auto mt-14 grid max-w-4xl grid-cols-2 gap-3 sm:mt-16 sm:grid-cols-4 sm:gap-4">
      {products.map(([src, alt], index) => (
        <div
          key={src}
          className={`relative aspect-[4/5] overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm ${
            index % 2 === 0 ? 'sm:translate-y-5' : ''
          }`}
        >
          <Image
            src={src}
            alt={alt}
            fill
            sizes="(max-width: 640px) 45vw, 220px"
            className="object-cover"
            priority={index < 2}
          />
        </div>
      ))}
    </FadeIn>
  )
}

function HowItWorks() {
  const steps = [
    {
      icon: <StoreIcon />,
      number: '01',
      title: 'Create your store',
      description: 'Add products, set prices in MVR, and get a shareable link in under 2 minutes.',
    },
    {
      icon: <QrIcon />,
      number: '02',
      title: 'Accept Swipe payments',
      description: 'Customers pay via QR code. Swipe confirms instantly - no bank transfers, no chasing.',
    },
    {
      icon: <ExchangeIcon />,
      number: '03',
      title: 'Trade or cash out',
      description: 'Convert earnings to USDT on the built-in P2P exchange, or keep MVR - your choice.',
    },
  ]

  return (
    <section className="bg-white">
      <div className="mx-auto max-w-6xl px-5 py-20 sm:px-8 lg:py-28">
        <FadeIn className="text-center">
          <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">How it works</p>
          <h2 className="mx-auto mt-4 max-w-2xl font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
            Three steps from post to payment.
          </h2>
        </FadeIn>

        <div className="mt-14 grid gap-8 sm:mt-16 md:grid-cols-3 md:gap-6">
          {steps.map((step, i) => (
            <FadeIn key={step.number} delay={i * 0.08}>
              <div className="relative flex flex-col items-start rounded-2xl border border-slate-100 bg-slate-50/50 p-7 sm:p-8">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-amethyst-100 text-amethyst-700">
                  {step.icon}
                </div>
                <span className="mt-5 font-mono text-xs text-slate-400">{step.number}</span>
                <h3 className="mt-1 font-display text-xl font-semibold text-slate-900">{step.title}</h3>
                <p className="mt-3 text-sm leading-6 text-slate-500">{step.description}</p>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  )
}

function BeforeAfter() {
  const manual = ['DM for price', 'Ask for bank transfer', 'Wait for slip', 'Manually reconcile']
  const swipe = ['Customer checks out', 'Swipe confirms payment', 'Order updates live', 'Seller ships faster']

  return (
    <section className="border-y border-slate-200/70 bg-slate-50">
      <div className="mx-auto grid max-w-6xl gap-10 px-5 py-20 sm:px-8 lg:grid-cols-[0.9fr_1.1fr] lg:items-center lg:py-28">
        <FadeIn>
          <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">Manual to automatic</p>
          <h2 className="mt-4 max-w-lg font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
            The social seller workflow, without the payment chase.
          </h2>
          <p className="mt-5 max-w-md text-base leading-7 text-slate-500">
            SwiftStore turns Instagram and Facebook demand into a real storefront with Swipe handling QR payment, confirmation, and reconciliation.
          </p>
        </FadeIn>

        <FadeIn className="grid gap-4 sm:grid-cols-2">
          <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <p className="font-display text-lg font-semibold text-slate-900">Before</p>
            <div className="mt-5 space-y-3">
              {manual.map((item) => (
                <div key={item} className="flex items-center gap-3 rounded-lg bg-slate-50 px-3 py-3 text-sm text-slate-500">
                  <span className="h-2 w-2 shrink-0 rounded-full bg-ruby-500" />
                  {item}
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border border-amethyst-200 bg-white p-5 shadow-md">
            <p className="font-display text-lg font-semibold text-slate-900">With Swipe</p>
            <div className="mt-5 space-y-3">
              {swipe.map((item) => (
                <div key={item} className="flex items-center gap-3 rounded-lg bg-amethyst-50 px-3 py-3 text-sm font-medium text-amethyst-800">
                  <span className="h-2 w-2 shrink-0 rounded-full bg-success" />
                  {item}
                </div>
              ))}
            </div>
          </div>
        </FadeIn>
      </div>
    </section>
  )
}

function Features() {
  const features = [
    {
      title: 'Storefronts',
      description: 'Product pages, checkout, order tracking, and shareable links - ready for Instagram and Facebook.',
      accent: 'bg-amethyst-600',
    },
    {
      title: 'Swipe Settlement',
      description: 'QR payments settle in MVR with live status updates. No more screenshot receipts.',
      accent: 'bg-success',
    },
    {
      title: 'P2P Exchange',
      description: 'A local exchange table lets buyers compare USDT rates and complete escrow-backed trades.',
      accent: 'bg-slate-900',
    },
  ]

  return (
    <section className="bg-white">
      <div className="mx-auto max-w-6xl px-5 py-20 sm:px-8 lg:py-28">
        <FadeIn className="mx-auto max-w-2xl text-center">
          <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-600">Built for demo day and real sellers</p>
          <h2 className="mt-4 font-display text-3xl font-semibold tracking-[-0.03em] text-slate-900 sm:text-5xl">
            Marketplace, payments, and P2P exchange in one flow.
          </h2>
        </FadeIn>

        <div className="mt-14 grid gap-6 sm:mt-16 md:grid-cols-3">
          {features.map((feature, i) => (
            <FadeIn key={feature.title} delay={i * 0.08}>
              <div className="group relative h-full overflow-hidden rounded-2xl border border-slate-200/80 bg-white p-7 shadow-sm transition-shadow hover:shadow-md sm:p-8">
                <div className={`h-1.5 w-10 rounded-full ${feature.accent}`} />
                <h3 className="mt-5 font-display text-xl font-semibold text-slate-900">{feature.title}</h3>
                <p className="mt-3 text-sm leading-6 text-slate-500">{feature.description}</p>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  )
}

/* ── Stats bar ── */
function StatsBar() {
  const stats = [
    { value: '< 2 min', label: 'Store setup time' },
    { value: 'Instant', label: 'Swipe confirmation' },
    { value: '0%', label: 'Platform fees (hackathon)' },
  ]

  return (
    <section className="border-y border-slate-200/70 bg-slate-50">
      <div className="mx-auto max-w-6xl px-5 py-12 sm:px-8 sm:py-14">
        <div className="grid grid-cols-3 gap-4 text-center">
          {stats.map((stat) => (
            <FadeIn key={stat.label}>
              <p className="font-display text-2xl font-bold tracking-tight text-slate-900 sm:text-4xl">{stat.value}</p>
              <p className="mt-1 text-xs font-medium text-slate-400 sm:text-sm">{stat.label}</p>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  )
}

/* ── Page ── */

export default function Home() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

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
        {/* Hero */}
        <section className="relative overflow-hidden">
          {/* Subtle gradient background */}
          <div className="absolute inset-0 bg-gradient-to-b from-amethyst-50/40 via-white to-white" aria-hidden="true" />

          <div className="relative mx-auto max-w-5xl px-5 pb-20 pt-20 text-center sm:px-8 sm:pb-24 sm:pt-28 lg:pb-28 lg:pt-32">
            {/* Live status pill */}
            <FadeIn>
              <div className="mb-8 inline-flex items-center gap-2.5 rounded-full border border-slate-200/80 bg-white px-4 py-2 shadow-sm">
                <span className="relative flex h-2 w-2">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75" />
                  <span className="relative inline-flex h-2 w-2 rounded-full bg-success" />
                </span>
                <span className="text-xs font-semibold text-slate-600">Live on Swipe - Maldives</span>
              </div>
            </FadeIn>

            <FadeIn delay={0.05}>
              <h1 className="mx-auto w-full max-w-[21rem] font-display text-[38px] font-semibold leading-[1] tracking-[-0.035em] text-slate-950 sm:max-w-3xl sm:text-[64px] sm:leading-[0.96] lg:text-[80px]">
                <span className="sm:hidden">
                  The selling layer
                  <br />
                  for Maldives
                  <br />
                  social commerce.
                </span>
                <span className="hidden sm:inline">The selling layer for Maldives social commerce.</span>
              </h1>
            </FadeIn>

            <FadeIn delay={0.1}>
              <p className="mx-auto mt-7 w-full max-w-[21rem] text-base leading-7 text-slate-500 sm:max-w-xl sm:text-lg sm:leading-8">
                Turn posts into storefronts, accept instant Swipe payments, and let buyers trade USDT - all from one platform.
              </p>
            </FadeIn>

            <FadeIn delay={0.15}>
              <div className="mt-10 flex w-full flex-col items-center justify-center gap-3 sm:w-auto sm:flex-row">
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
              </div>
            </FadeIn>

            <ProductStrip />
          </div>
        </section>

        <StatsBar />
        <HowItWorks />
        <BeforeAfter />
        <Features />

        {/* Final CTA */}
        <section className="bg-slate-900 px-5 py-20 text-center sm:px-8 lg:py-28">
          <FadeIn className="mx-auto max-w-3xl">
            <p className="text-xs font-bold uppercase tracking-[0.15em] text-amethyst-300">Ready for sellers</p>
            <h2 className="mt-4 font-display text-3xl font-semibold tracking-[-0.03em] text-white sm:text-5xl">
              Launch a store, take payment, and prove the order in minutes.
            </h2>
            <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Link
                href="/seller/login"
                className="inline-flex w-full items-center justify-center rounded-xl bg-white px-8 py-4 text-sm font-semibold text-slate-900 transition-colors hover:bg-slate-100 sm:w-auto"
              >
                Open Seller Dashboard
              </Link>
              <Link
                href="/exchange"
                className="inline-flex w-full items-center justify-center rounded-xl border border-slate-600 px-8 py-4 text-sm font-semibold text-slate-300 transition-colors hover:border-slate-500 hover:text-white sm:w-auto"
              >
                Explore Exchange
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
          <p className="text-xs text-slate-400">&copy; 2026 SwiftStore &middot; Powered by Swipe</p>
        </div>
      </footer>
    </div>
  )
}
