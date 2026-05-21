'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { useIsMobile } from '@/lib/use-device'
import { useState } from 'react'

/* ---- Fade-in wrapper (opacity only, no movement) ---- */
function FadeIn({ children, className = '' }: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true, margin: '-40px' }}
      transition={{ duration: 0.3, ease: 'easeOut' }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

/* ======== PAGE ======== */
export default function Home() {
  const isMobile = useIsMobile()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  if (isMobile === null) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="w-8 h-8 rounded-lg bg-amethyst-500 animate-pulse" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white font-body text-slate-900">

      {/* ---- NAV ---- */}
      <nav className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-[rgba(15,23,42,0.06)]">
        <div className={`mx-auto max-w-6xl flex items-center justify-between ${isMobile ? 'px-5 h-14' : 'px-8 h-16'}`}>
          <Link href="/" className="flex items-center gap-2">
            <span className="font-display font-bold text-slate-900 text-lg tracking-tight">SwiftStore</span>
          </Link>

          {!isMobile && (
            <div className="flex items-center gap-8">
              <Link href="/marketplace" className="text-sm text-slate-500 hover:text-slate-900 transition-colors font-medium">
                Marketplace
              </Link>
              <Link href="/exchange" className="text-sm text-slate-500 hover:text-slate-900 transition-colors font-medium">
                Exchange
              </Link>
              <Link
                href="/seller/login"
                className="text-sm bg-amethyst-500 text-white px-4 py-2 rounded-lg font-semibold hover:bg-amethyst-600 transition-colors"
              >
                Start Selling
              </Link>
            </div>
          )}

          {isMobile && (
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="text-slate-500 p-1"
              aria-label="Menu"
            >
              <svg className="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                {mobileMenuOpen ? (
                  <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                ) : (
                  <path fillRule="evenodd" d="M3 5a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 15a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
                )}
              </svg>
            </button>
          )}
        </div>

        {/* Mobile menu */}
        {isMobile && mobileMenuOpen && (
          <div className="border-t border-[rgba(15,23,42,0.06)] bg-white px-5 py-4 space-y-3">
            <Link href="/marketplace" className="block text-sm text-slate-600 font-medium" onClick={() => setMobileMenuOpen(false)}>
              Marketplace
            </Link>
            <Link href="/exchange" className="block text-sm text-slate-600 font-medium" onClick={() => setMobileMenuOpen(false)}>
              Exchange
            </Link>
            <Link
              href="/seller/login"
              className="block text-sm bg-amethyst-500 text-white px-4 py-2 rounded-lg font-semibold text-center"
              onClick={() => setMobileMenuOpen(false)}
            >
              Start Selling
            </Link>
          </div>
        )}
      </nav>

      {/* ---- HERO ---- */}
      <section className={isMobile ? 'px-5 pt-16 pb-20' : 'px-8 pt-32 pb-32'}>
        <div className="mx-auto max-w-3xl text-center">
          <FadeIn>
            <h1 className={`font-display font-bold tracking-tighter leading-[1.05] text-slate-900 ${isMobile ? 'text-[32px]' : 'text-[56px]'}`}>
              Commerce and crypto,
              <br />
              <span className="text-amethyst-500">one platform.</span>
            </h1>
          </FadeIn>

          <FadeIn>
            <p className={`mt-6 text-slate-500 leading-relaxed max-w-lg mx-auto ${isMobile ? 'text-[15px]' : 'text-lg'}`}>
              A social marketplace and P2P crypto exchange for the Maldives. Buy products, trade USDT, settle in MVR with Swipe.
            </p>
          </FadeIn>

          <FadeIn>
            <div className={`mt-10 flex items-center justify-center gap-3 ${isMobile ? 'flex-col' : ''}`}>
              <Link
                href="/marketplace"
                className={`inline-flex items-center justify-center rounded-lg bg-amethyst-500 text-white font-semibold hover:bg-amethyst-600 transition-colors ${isMobile ? 'w-full px-6 py-3 text-sm' : 'px-6 py-3 text-sm'}`}
              >
                Browse Marketplace
              </Link>
              <Link
                href="/exchange"
                className={`inline-flex items-center justify-center rounded-lg border border-[rgba(15,23,42,0.12)] text-slate-700 font-semibold hover:bg-slate-50 transition-colors ${isMobile ? 'w-full px-6 py-3 text-sm' : 'px-6 py-3 text-sm'}`}
              >
                Trade Crypto
              </Link>
            </div>
          </FadeIn>
        </div>
      </section>

      {/* ---- FEATURES ---- */}
      <section className={isMobile ? 'px-5 pb-20' : 'px-8 pb-32'}>
        <div className={`mx-auto max-w-4xl ${isMobile ? 'flex flex-col gap-4' : 'grid grid-cols-2 gap-6'}`}>
          {/* Social Commerce */}
          <FadeIn>
            <div className="rounded-xl border border-[rgba(15,23,42,0.06)] p-8 h-full">
              <h3 className={`font-display font-bold tracking-tight text-slate-900 ${isMobile ? 'text-lg' : 'text-xl'}`}>
                Social Commerce
              </h3>
              <p className="mt-3 text-sm text-slate-500 leading-relaxed">
                Sellers from Instagram and Facebook get a real storefront. Customers browse, checkout, and pay with Swipe.
              </p>
              <ul className="mt-5 space-y-2.5">
                {['One-tap checkout with Swipe', 'Auto-sync from social media', 'Real-time order tracking'].map((item) => (
                  <li key={item} className="flex items-center gap-2.5 text-sm text-slate-600">
                    <span className="w-1 h-1 rounded-full bg-amethyst-500 flex-shrink-0" />
                    {item}
                  </li>
                ))}
              </ul>
              <Link
                href="/marketplace"
                className="inline-block mt-6 text-sm font-medium text-amethyst-500 hover:text-amethyst-600 transition-colors"
              >
                Explore marketplace
              </Link>
            </div>
          </FadeIn>

          {/* P2P Exchange */}
          <FadeIn>
            <div className="rounded-xl border border-[rgba(15,23,42,0.06)] p-8 h-full">
              <h3 className={`font-display font-bold tracking-tight text-slate-900 ${isMobile ? 'text-lg' : 'text-xl'}`}>
                P2P Exchange
              </h3>
              <p className="mt-3 text-sm text-slate-500 leading-relaxed">
                Buy and sell USDT peer-to-peer. Escrow holds the crypto until Swipe payment confirms. No middlemen.
              </p>
              <ul className="mt-5 space-y-2.5">
                {['Buy & sell USDT with MVR', 'Escrow-protected trades', 'Competitive local rates'].map((item) => (
                  <li key={item} className="flex items-center gap-2.5 text-sm text-slate-600">
                    <span className="w-1 h-1 rounded-full bg-amethyst-500 flex-shrink-0" />
                    {item}
                  </li>
                ))}
              </ul>
              <Link
                href="/exchange"
                className="inline-block mt-6 text-sm font-medium text-amethyst-500 hover:text-amethyst-600 transition-colors"
              >
                Start trading
              </Link>
            </div>
          </FadeIn>
        </div>
      </section>

      {/* ---- HOW IT WORKS ---- */}
      <section className={`bg-slate-50 ${isMobile ? 'px-5 py-20' : 'px-8 py-32'}`}>
        <div className="mx-auto max-w-4xl">
          <FadeIn>
            <h2 className={`font-display font-bold tracking-tight text-slate-900 text-center ${isMobile ? 'text-2xl' : 'text-3xl'}`}>
              How it works
            </h2>
          </FadeIn>

          <div className={`mt-14 ${isMobile ? 'flex flex-col gap-10' : 'grid grid-cols-3 gap-12'}`}>
            {[
              { num: '01', title: 'List or browse', desc: 'Sellers list products and USDT offers. Buyers find what they need.' },
              { num: '02', title: 'Pay with Swipe', desc: 'Instant MVR settlement. No bank transfers, no screenshots.' },
              { num: '03', title: 'Receive', desc: 'Products ship to your door. USDT transfers to your wallet. Done.' },
            ].map((step) => (
              <FadeIn key={step.num}>
                <div className={isMobile ? 'text-center' : ''}>
                  <span className="font-mono text-sm text-amethyst-500 font-medium">{step.num}</span>
                  <h3 className="font-display font-bold text-slate-900 text-lg tracking-tight mt-2">{step.title}</h3>
                  <p className="mt-2 text-sm text-slate-500 leading-relaxed">{step.desc}</p>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ---- FINAL CTA ---- */}
      <section className={`bg-slate-900 ${isMobile ? 'px-5 py-20' : 'px-8 py-32'}`}>
        <div className="mx-auto max-w-3xl text-center">
          <FadeIn>
            <h2 className={`font-display font-bold tracking-tight text-white ${isMobile ? 'text-2xl' : 'text-3xl'}`}>
              Start selling today.
            </h2>
            <p className={`mt-4 text-slate-400 ${isMobile ? 'text-sm' : 'text-base'}`}>
              Set up your store in minutes. Reach customers across the Maldives.
            </p>
            <Link
              href="/seller/login"
              className="inline-flex items-center justify-center mt-8 rounded-lg bg-ruby-500 text-white px-6 py-3 text-sm font-semibold hover:bg-ruby-600 transition-colors"
            >
              Start Selling
            </Link>
          </FadeIn>
        </div>
      </section>

      {/* ---- FOOTER ---- */}
      <footer className="border-t border-[rgba(15,23,42,0.06)]">
        <div className={`mx-auto max-w-6xl ${isMobile ? 'px-5 py-8' : 'px-8 py-10'}`}>
          <div className={`${isMobile ? 'text-center space-y-3' : 'flex items-center justify-between'}`}>
            <span className="font-display font-bold text-slate-900 text-sm tracking-tight">SwiftStore</span>
            <div className={`flex items-center gap-6 ${isMobile ? 'justify-center' : ''}`}>
              <Link href="/marketplace" className="text-xs text-slate-400 hover:text-slate-600 transition-colors">Marketplace</Link>
              <Link href="/exchange" className="text-xs text-slate-400 hover:text-slate-600 transition-colors">Exchange</Link>
              <Link href="/seller/login" className="text-xs text-slate-400 hover:text-slate-600 transition-colors">Sellers</Link>
            </div>
            <p className={`text-xs text-slate-400 ${isMobile ? '' : ''}`}>
              &copy; 2026 SwiftStore &middot; Powered by Swipe
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
