'use client'

import Link from 'next/link'
import { useRef } from 'react'
import { motion, useInView, useMotionValue, useTransform, animate } from 'framer-motion'
import { useIsMobile } from '@/lib/use-device'
import { useEffect, useState } from 'react'

/* ---- Animated Counter ---- */
function AnimatedCounter({ target, suffix = '', prefix = '' }: { target: number; suffix?: string; prefix?: string }) {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: '-50px' })
  const [display, setDisplay] = useState(0)

  useEffect(() => {
    if (!isInView) return
    const controls = animate(0, target, {
      duration: 2,
      ease: 'easeOut',
      onUpdate(value) {
        setDisplay(Math.floor(value))
      },
    })
    return () => controls.stop()
  }, [isInView, target])

  return (
    <span ref={ref} className="font-mono tabular-nums">
      {prefix}{display.toLocaleString()}{suffix}
    </span>
  )
}

/* ---- Fade-in wrapper ---- */
function FadeIn({ children, delay = 0, className = '', direction = 'up' }: {
  children: React.ReactNode
  delay?: number
  className?: string
  direction?: 'up' | 'down' | 'left' | 'right' | 'none'
}) {
  const directionMap = {
    up: { y: 40, x: 0 },
    down: { y: -40, x: 0 },
    left: { x: 40, y: 0 },
    right: { x: -40, y: 0 },
    none: { x: 0, y: 0 },
  }
  const d = directionMap[direction]
  return (
    <motion.div
      initial={{ opacity: 0, x: d.x, y: d.y }}
      whileInView={{ opacity: 1, x: 0, y: 0 }}
      viewport={{ once: true, margin: '-60px' }}
      transition={{ duration: 0.7, delay, ease: [0.16, 1, 0.3, 1] }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

/* ---- Floating Orb (background decoration) ---- */
function FloatingOrb({ className }: { className: string }) {
  return (
    <motion.div
      className={`absolute rounded-full blur-3xl opacity-20 ${className}`}
      animate={{
        y: [0, -20, 0],
        scale: [1, 1.05, 1],
      }}
      transition={{
        duration: 6,
        repeat: Infinity,
        ease: 'easeInOut',
      }}
    />
  )
}

/* ---- Grid Pattern Background ---- */
function GridPattern() {
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      <svg className="absolute inset-0 w-full h-full" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="grid" width="60" height="60" patternUnits="userSpaceOnUse">
            <path d="M 60 0 L 0 0 0 60" fill="none" stroke="currentColor" strokeWidth="0.5" className="text-white/[0.07]" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
      </svg>
    </div>
  )
}

/* ---- Phone Mockup (CSS-only, premium) ---- */
function PhoneMockup() {
  return (
    <motion.div
      className="relative mx-auto w-[280px] h-[520px]"
      initial={{ opacity: 0, y: 60, rotateY: -8 }}
      animate={{ opacity: 1, y: 0, rotateY: 0 }}
      transition={{ duration: 1, delay: 0.4, ease: [0.16, 1, 0.3, 1] }}
    >
      {/* Glow behind phone */}
      <div className="absolute -inset-8 bg-gradient-to-b from-amethyst-500/30 to-ruby-500/20 rounded-[3rem] blur-2xl" />

      <div className="relative w-full h-full rounded-[2.5rem] border border-white/20 bg-white/95 backdrop-blur-xl shadow-2xl overflow-hidden">
        {/* Notch */}
        <div className="absolute top-0 inset-x-0 flex justify-center">
          <div className="w-28 h-6 bg-slate-900 rounded-b-2xl" />
        </div>
        {/* Status bar */}
        <div className="h-10 flex items-end justify-between px-7 pb-1">
          <span className="text-[10px] text-slate-400 font-semibold">9:41</span>
          <div className="flex gap-1 items-center">
            <div className="w-3.5 h-2 rounded-sm border border-slate-300" />
          </div>
        </div>
        {/* App header */}
        <div className="px-5 pt-2 pb-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-amethyst-500 to-amethyst-700 flex items-center justify-center shadow-md">
              <span className="text-white text-xs font-bold">S</span>
            </div>
            <div>
              <span className="text-xs font-bold text-slate-800 block leading-tight">SwipeStore</span>
              <span className="text-[9px] text-slate-400">Marketplace</span>
            </div>
          </div>
          <div className="w-7 h-7 rounded-full bg-slate-100 flex items-center justify-center">
            <div className="w-3 h-3 rounded-full bg-slate-300" />
          </div>
        </div>

        {/* Balance card */}
        <div className="mx-4 rounded-xl bg-gradient-to-r from-amethyst-600 to-amethyst-800 p-3.5 mb-3 shadow-lg">
          <p className="text-[9px] text-amethyst-200 font-medium">Swipe Balance</p>
          <p className="text-lg font-bold text-white font-mono mt-0.5">MVR 12,450</p>
          <div className="flex gap-2 mt-2">
            <div className="flex-1 rounded-lg bg-white/20 py-1.5 text-center">
              <span className="text-[9px] text-white font-semibold">Send</span>
            </div>
            <div className="flex-1 rounded-lg bg-white/20 py-1.5 text-center">
              <span className="text-[9px] text-white font-semibold">Top Up</span>
            </div>
          </div>
        </div>

        {/* Exchange listing */}
        <div className="mx-4 mb-2.5 rounded-xl border border-slate-100 p-3 bg-white">
          <div className="flex items-center justify-between mb-1.5">
            <div className="flex items-center gap-1.5">
              <div className="w-5 h-5 rounded-full bg-green-100 flex items-center justify-center">
                <span className="text-[8px] font-bold text-green-700">$</span>
              </div>
              <span className="text-[11px] font-bold text-slate-800">100 USDT</span>
            </div>
            <span className="text-[8px] font-semibold px-2 py-0.5 rounded-full bg-green-50 text-green-600 border border-green-100">Escrowed</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-[9px] text-slate-400">25.50 MVR/USDT</span>
            <span className="text-[11px] font-bold text-amethyst-600 font-mono">2,550 MVR</span>
          </div>
        </div>

        {/* Product card */}
        <div className="mx-4 mb-3 rounded-xl border border-slate-100 p-3 bg-white">
          <div className="flex gap-3">
            <div className="w-14 h-14 rounded-lg bg-gradient-to-br from-amethyst-50 to-ruby-50 flex-shrink-0 flex items-center justify-center">
              <div className="w-8 h-10 rounded bg-gradient-to-b from-slate-200 to-slate-300" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-[11px] font-bold text-slate-800">Black Abaya</div>
              <div className="text-[9px] text-slate-400">Island Finds MV</div>
              <div className="flex items-center justify-between mt-1.5">
                <div className="text-[11px] font-bold text-amethyst-600 font-mono">MVR 650</div>
                <div className="rounded-md bg-amethyst-600 px-2.5 py-1">
                  <span className="text-[8px] font-semibold text-white">Buy Now</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom nav */}
        <div className="absolute bottom-0 inset-x-0 h-14 border-t border-slate-100 bg-white/90 backdrop-blur flex items-center justify-around px-8">
          <div className="flex flex-col items-center gap-0.5">
            <div className="w-5 h-5 rounded bg-amethyst-100 flex items-center justify-center">
              <div className="w-2.5 h-2.5 rounded-sm bg-amethyst-500" />
            </div>
            <span className="text-[7px] font-semibold text-amethyst-600">Shop</span>
          </div>
          <div className="flex flex-col items-center gap-0.5">
            <div className="w-5 h-5 rounded bg-slate-100 flex items-center justify-center">
              <div className="w-2.5 h-2.5 rounded-full bg-slate-300" />
            </div>
            <span className="text-[7px] text-slate-400">Exchange</span>
          </div>
          <div className="flex flex-col items-center gap-0.5">
            <div className="w-5 h-5 rounded bg-slate-100 flex items-center justify-center">
              <div className="w-2.5 h-2.5 rounded-sm bg-slate-300" />
            </div>
            <span className="text-[7px] text-slate-400">Profile</span>
          </div>
        </div>
      </div>
    </motion.div>
  )
}

/* ---- Arrow Icon ---- */
function ArrowRight({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clipRule="evenodd" />
    </svg>
  )
}

/* ---- Feature Icons (inline SVG) ---- */
function ShoppingIcon() {
  return (
    <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z" />
      <line x1="3" y1="6" x2="21" y2="6" />
      <path d="M16 10a4 4 0 01-8 0" />
    </svg>
  )
}

function ExchangeIcon() {
  return (
    <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="17 1 21 5 17 9" />
      <path d="M3 11V9a4 4 0 014-4h14" />
      <polyline points="7 23 3 19 7 15" />
      <path d="M21 13v2a4 4 0 01-4 4H3" />
    </svg>
  )
}

function ShieldIcon() {
  return (
    <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <polyline points="9 12 11 14 15 10" />
    </svg>
  )
}

function ZapIcon() {
  return (
    <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
    </svg>
  )
}

/* ======== PAGE ======== */
export default function Home() {
  const isMobile = useIsMobile()

  if (isMobile === null) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <motion.div
          className="w-10 h-10 rounded-xl bg-gradient-to-br from-amethyst-500 to-ruby-500"
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
        />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white font-body text-slate-900 overflow-x-hidden">

      {/* ---- NAV BAR ---- */}
      <motion.nav
        className="sticky top-0 z-50 bg-white/70 backdrop-blur-xl border-b border-slate-100/50"
        initial={{ y: -80 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
      >
        <div className={`mx-auto max-w-7xl flex items-center justify-between ${isMobile ? 'px-5 h-14' : 'px-8 h-16'}`}>
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amethyst-500 to-amethyst-700 flex items-center justify-center shadow-md group-hover:shadow-lg transition-shadow">
              <span className="text-white text-sm font-bold">S</span>
            </div>
            <span className="font-display font-extrabold text-slate-900 text-lg tracking-tight">SwipeStore</span>
          </Link>
          {!isMobile && (
            <div className="flex items-center gap-1">
              <Link href="/marketplace" className="text-sm text-slate-500 hover:text-slate-900 transition-colors font-medium px-4 py-2 rounded-lg hover:bg-slate-50">Marketplace</Link>
              <Link href="/exchange" className="text-sm text-slate-500 hover:text-slate-900 transition-colors font-medium px-4 py-2 rounded-lg hover:bg-slate-50">Exchange</Link>
              <div className="w-px h-6 bg-slate-200 mx-2" />
              <Link href="/seller/login" className="text-sm bg-gradient-to-r from-amethyst-600 to-amethyst-700 text-white px-5 py-2.5 rounded-xl font-semibold hover:shadow-lg hover:shadow-amethyst-500/25 transition-all">
                Start Selling
              </Link>
            </div>
          )}
          {isMobile && (
            <Link href="/seller/login" className="text-xs bg-gradient-to-r from-amethyst-600 to-amethyst-700 text-white px-4 py-2 rounded-lg font-semibold">
              Start Selling
            </Link>
          )}
        </div>
      </motion.nav>

      {/* ---- HERO SECTION ---- */}
      <section className="relative overflow-hidden">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-900 to-amethyst-900" />
        <GridPattern />

        {/* Floating orbs */}
        <FloatingOrb className="w-96 h-96 bg-amethyst-500 -top-20 -right-20" />
        <FloatingOrb className="w-72 h-72 bg-ruby-500 bottom-10 -left-20" />
        <FloatingOrb className="w-64 h-64 bg-amethyst-400 top-1/2 left-1/3" />

        <div className={`relative mx-auto max-w-7xl ${isMobile ? 'px-5 pt-14 pb-16' : 'px-8 pt-24 pb-32'}`}>
          <div className={`${isMobile ? '' : 'flex items-center gap-20'}`}>
            {/* Text side */}
            <div className={`${isMobile ? '' : 'flex-1'}`}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.1 }}
                className="inline-flex items-center gap-2 rounded-full bg-white/10 border border-white/10 backdrop-blur-sm px-4 py-2 mb-6"
              >
                <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
                <span className="text-xs font-semibold text-white/80 tracking-wide">Live in the Maldives</span>
              </motion.div>

              <motion.h1
                className={`font-display font-extrabold tracking-tight text-white leading-[1.08] ${isMobile ? 'text-[32px]' : 'text-[56px]'}`}
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
              >
                Shop, Sell &{' '}
                <span className="bg-gradient-to-r from-amethyst-400 via-ruby-400 to-amethyst-400 bg-clip-text text-transparent">
                  Trade Crypto
                </span>
                {' '}on One Platform
              </motion.h1>

              <motion.p
                className={`mt-5 text-slate-300 leading-relaxed ${isMobile ? 'text-[15px]' : 'text-lg max-w-xl'}`}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, delay: 0.35 }}
              >
                SwipeStore brings social commerce and P2P crypto exchange together.
                Buy products from local sellers. Buy and sell USDT with escrow protection.
                All powered by instant Swipe payments.
              </motion.p>

              {/* CTAs */}
              <motion.div
                className={`mt-8 flex gap-3 ${isMobile ? 'flex-col' : 'flex-row'}`}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, delay: 0.5 }}
              >
                <Link
                  href="/marketplace"
                  className="group inline-flex items-center justify-center gap-2.5 rounded-xl bg-white px-7 py-3.5 text-sm font-bold text-slate-900 shadow-xl hover:shadow-2xl transition-all hover:scale-[1.02] active:scale-[0.98]"
                >
                  Browse Marketplace
                  <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                </Link>
                <Link
                  href="/exchange"
                  className="group inline-flex items-center justify-center gap-2.5 rounded-xl bg-gradient-to-r from-amethyst-500 to-ruby-500 px-7 py-3.5 text-sm font-bold text-white shadow-xl shadow-amethyst-500/25 hover:shadow-2xl hover:shadow-amethyst-500/30 transition-all hover:scale-[1.02] active:scale-[0.98]"
                >
                  Trade Crypto
                  <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                </Link>
              </motion.div>

              {/* Trust indicators */}
              <motion.div
                className={`mt-8 flex items-center gap-6 ${isMobile ? 'flex-wrap gap-4' : ''}`}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.7, delay: 0.7 }}
              >
                {[
                  'Escrow Protected',
                  'Instant Payments',
                  'Zero Scams',
                ].map((item) => (
                  <div key={item} className="flex items-center gap-2">
                    <svg className="w-4 h-4 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span className="text-xs text-slate-400 font-medium">{item}</span>
                  </div>
                ))}
              </motion.div>
            </div>

            {/* Phone mockup */}
            {!isMobile && (
              <div className="flex-shrink-0 perspective-[1200px]">
                <PhoneMockup />
              </div>
            )}
          </div>
        </div>

        {/* Gradient fade to white */}
        <div className="absolute bottom-0 left-0 right-0 h-24 bg-gradient-to-t from-white to-transparent" />
      </section>

      {/* ---- DUAL FEATURE SHOWCASE ---- */}
      <section className={`relative ${isMobile ? 'px-5 py-16' : 'px-8 py-24'}`}>
        <div className="mx-auto max-w-7xl">
          <FadeIn className="text-center mb-16">
            <p className="text-sm font-semibold text-amethyst-600 tracking-wide uppercase mb-3">Two Platforms, One App</p>
            <h2 className={`font-display font-extrabold text-slate-900 ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
              Everything you need to
              <br className={isMobile ? 'hidden' : ''} />
              {' '}buy, sell, and trade
            </h2>
          </FadeIn>

          <div className={`${isMobile ? 'flex flex-col gap-6' : 'grid grid-cols-2 gap-8'}`}>
            {/* Marketplace Card */}
            <FadeIn delay={0.1} direction="left">
              <div className="group relative rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-amethyst-50/30 p-8 hover:border-amethyst-200 hover:shadow-xl hover:shadow-amethyst-500/5 transition-all duration-500 overflow-hidden">
                {/* Decorative gradient */}
                <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-amethyst-100/50 to-transparent rounded-bl-full" />

                <div className="relative">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-amethyst-500 to-amethyst-700 flex items-center justify-center mb-6 shadow-lg shadow-amethyst-500/20 group-hover:scale-110 transition-transform">
                    <ShoppingIcon />
                  </div>
                  <h3 className={`font-display font-bold text-slate-900 ${isMobile ? 'text-xl' : 'text-2xl'}`}>Social Marketplace</h3>
                  <p className="mt-3 text-slate-500 leading-relaxed text-[15px]">
                    Instagram and Facebook sellers get their own storefront. Customers browse, buy, and pay — all with Swipe. No more DM negotiations.
                  </p>

                  <div className="mt-6 space-y-3">
                    {[
                      'One-tap checkout with Swipe',
                      'Auto-sync from social media',
                      'Real-time order tracking',
                      'Seller analytics dashboard',
                    ].map((feature, i) => (
                      <div key={i} className="flex items-center gap-3">
                        <div className="w-5 h-5 rounded-full bg-amethyst-100 flex items-center justify-center flex-shrink-0">
                          <svg className="w-3 h-3 text-amethyst-600" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        </div>
                        <span className="text-sm text-slate-600">{feature}</span>
                      </div>
                    ))}
                  </div>

                  <Link
                    href="/marketplace"
                    className="inline-flex items-center gap-2 mt-8 text-sm font-semibold text-amethyst-600 hover:text-amethyst-700 transition-colors group/link"
                  >
                    Explore Marketplace
                    <ArrowRight className="w-4 h-4 group-hover/link:translate-x-1 transition-transform" />
                  </Link>
                </div>
              </div>
            </FadeIn>

            {/* Exchange Card */}
            <FadeIn delay={0.2} direction="right">
              <div className="group relative rounded-2xl border border-slate-700 bg-gradient-to-br from-slate-800 to-slate-900 p-8 hover:border-slate-600 hover:shadow-xl hover:shadow-amethyst-500/10 transition-all duration-500 overflow-hidden">
                {/* Decorative gradient */}
                <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-amethyst-500/10 to-transparent rounded-bl-full" />

                <div className="relative">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-ruby-500 to-amethyst-600 flex items-center justify-center mb-6 shadow-lg shadow-ruby-500/20 group-hover:scale-110 transition-transform">
                    <ExchangeIcon />
                  </div>
                  <h3 className={`font-display font-bold text-white ${isMobile ? 'text-xl' : 'text-2xl'}`}>P2P Crypto Exchange</h3>
                  <p className="mt-3 text-slate-400 leading-relaxed text-[15px]">
                    Buy and sell USDT peer-to-peer with Swipe as the fiat rail. On-chain escrow protects every trade. No middlemen, no scams.
                  </p>

                  <div className="mt-6 space-y-3">
                    {[
                      'Buy & sell USDT instantly',
                      'On-chain escrow protection',
                      'Competitive MVR rates',
                      'Verified trader profiles',
                    ].map((feature, i) => (
                      <div key={i} className="flex items-center gap-3">
                        <div className="w-5 h-5 rounded-full bg-amethyst-500/20 flex items-center justify-center flex-shrink-0">
                          <svg className="w-3 h-3 text-amethyst-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        </div>
                        <span className="text-sm text-slate-300">{feature}</span>
                      </div>
                    ))}
                  </div>

                  <Link
                    href="/exchange"
                    className="inline-flex items-center gap-2 mt-8 text-sm font-semibold text-amethyst-400 hover:text-amethyst-300 transition-colors group/link"
                  >
                    Start Trading
                    <ArrowRight className="w-4 h-4 group-hover/link:translate-x-1 transition-transform" />
                  </Link>
                </div>
              </div>
            </FadeIn>
          </div>
        </div>
      </section>

      {/* ---- HOW IT WORKS ---- */}
      <section className={`relative bg-slate-50 ${isMobile ? 'px-5 py-16' : 'px-8 py-24'}`}>
        <div className="mx-auto max-w-6xl">
          <FadeIn className="text-center mb-16">
            <p className="text-sm font-semibold text-amethyst-600 tracking-wide uppercase mb-3">Simple by Design</p>
            <h2 className={`font-display font-extrabold text-slate-900 ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
              Three steps. That&apos;s it.
            </h2>
            <p className="mt-4 text-slate-500 max-w-lg mx-auto">
              Whether you&apos;re shopping for products or trading crypto, SwipeStore keeps it effortless.
            </p>
          </FadeIn>

          <div className={`${isMobile ? 'flex flex-col gap-8' : 'grid grid-cols-3 gap-10'}`}>
            {[
              {
                step: 1,
                title: 'Browse',
                desc: 'Find products from Maldivian sellers or USDT listings at the best rates. Everything in one place.',
                gradient: 'from-amethyst-500 to-amethyst-600',
              },
              {
                step: 2,
                title: 'Pay with Swipe',
                desc: 'Instant MVR payment — no bank transfers, no screenshots, no verification delays. Just tap and done.',
                gradient: 'from-amethyst-600 to-ruby-500',
              },
              {
                step: 3,
                title: 'Receive',
                desc: 'Products ship to your door. USDT transfers to your wallet. Automatic confirmation for both sides.',
                gradient: 'from-ruby-500 to-ruby-600',
              },
            ].map((item, i) => (
              <FadeIn key={item.step} delay={i * 0.15}>
                <div className="relative">
                  {/* Connector line (desktop only) */}
                  {!isMobile && i < 2 && (
                    <div className="absolute top-8 left-[calc(50%+2rem)] right-0 h-px bg-gradient-to-r from-slate-300 to-slate-200 -mr-10" />
                  )}
                  <div className="text-center">
                    <div className={`w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br ${item.gradient} flex items-center justify-center text-white text-xl font-bold shadow-lg shadow-amethyst-500/15 mb-5`}>
                      {item.step}
                    </div>
                    <h3 className="font-display font-bold text-slate-900 text-lg">{item.title}</h3>
                    <p className="mt-2.5 text-slate-500 text-sm leading-relaxed max-w-xs mx-auto">
                      {item.desc}
                    </p>
                  </div>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ---- STATS SECTION ---- */}
      <section className={`relative overflow-hidden ${isMobile ? 'px-5 py-16' : 'px-8 py-24'}`}>
        <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-amethyst-900/90 to-slate-900" />
        <GridPattern />
        <FloatingOrb className="w-80 h-80 bg-amethyst-500 -top-20 right-20" />
        <FloatingOrb className="w-60 h-60 bg-ruby-500 bottom-0 left-10" />

        <div className="relative mx-auto max-w-6xl">
          <FadeIn className="text-center mb-14">
            <p className="text-sm font-semibold text-amethyst-400 tracking-wide uppercase mb-3">Growing Fast</p>
            <h2 className={`font-display font-extrabold text-white ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
              Trusted by the Maldives
            </h2>
          </FadeIn>

          <div className={`${isMobile ? 'grid grid-cols-2 gap-6' : 'grid grid-cols-4 gap-8'}`}>
            {[
              { value: 2847, suffix: '+', label: 'Transactions', sublabel: 'processed' },
              { value: 156, suffix: '', label: 'Active Sellers', sublabel: 'onboarded' },
              { value: 425, suffix: 'K', label: 'MVR Volume', prefix: '', sublabel: 'monthly' },
              { value: 98, suffix: '%', label: 'Satisfaction', sublabel: 'rate' },
            ].map((stat, i) => (
              <FadeIn key={stat.label} delay={i * 0.1}>
                <div className="text-center rounded-2xl bg-white/5 backdrop-blur border border-white/10 p-6 hover:bg-white/10 transition-colors">
                  <div className={`font-display font-extrabold text-white ${isMobile ? 'text-3xl' : 'text-4xl'} mb-1`}>
                    <AnimatedCounter target={stat.value} suffix={stat.suffix} prefix={stat.prefix || ''} />
                  </div>
                  <p className="text-sm font-semibold text-slate-300">{stat.label}</p>
                  <p className="text-xs text-slate-500">{stat.sublabel}</p>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ---- WHY SWIPESTORE ---- */}
      <section className={`${isMobile ? 'px-5 py-16' : 'px-8 py-24'}`}>
        <div className="mx-auto max-w-6xl">
          <FadeIn className="text-center mb-14">
            <p className="text-sm font-semibold text-amethyst-600 tracking-wide uppercase mb-3">Why SwipeStore</p>
            <h2 className={`font-display font-extrabold text-slate-900 ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
              Built for how the Maldives
              <br className={isMobile ? 'hidden' : ''} /> actually does business
            </h2>
          </FadeIn>

          <div className={`${isMobile ? 'flex flex-col gap-5' : 'grid grid-cols-2 gap-6'}`}>
            {[
              {
                icon: <ShieldIcon />,
                title: 'Zero Scams',
                desc: 'Every transaction is verified by Swipe before products ship or crypto transfers. No fake screenshots, no chargebacks, no risk.',
                color: 'from-ruby-500 to-ruby-600',
                bg: 'bg-ruby-50',
              },
              {
                icon: <ZapIcon />,
                title: 'Instant Settlement',
                desc: 'Swipe confirms payment in seconds. Sellers see revenue immediately. Crypto releases the moment payment clears.',
                color: 'from-amethyst-500 to-amethyst-600',
                bg: 'bg-amethyst-50',
              },
              {
                icon: (
                  <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="2" y1="12" x2="22" y2="12" />
                    <path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
                  </svg>
                ),
                title: 'Made for Maldives',
                desc: 'MVR payments, Maldivian sellers, island delivery. Not a generic platform adapted for us — built from scratch for our market.',
                color: 'from-blue-500 to-blue-600',
                bg: 'bg-blue-50',
              },
              {
                icon: (
                  <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
                    <line x1="1" y1="10" x2="23" y2="10" />
                  </svg>
                ),
                title: 'Full Transparency',
                desc: 'Track every order and every trade. Escrow balances visible in real-time. Complete transaction history at your fingertips.',
                color: 'from-green-500 to-green-600',
                bg: 'bg-green-50',
              },
            ].map((feature, i) => (
              <FadeIn key={feature.title} delay={i * 0.1}>
                <div className="group rounded-2xl border border-slate-200 bg-white p-7 hover:border-amethyst-200 hover:shadow-lg transition-all duration-300">
                  <div className="flex items-start gap-5">
                    <div className={`w-12 h-12 rounded-xl ${feature.bg} flex items-center justify-center flex-shrink-0 text-slate-700 group-hover:scale-110 transition-transform`}>
                      {feature.icon}
                    </div>
                    <div>
                      <h3 className="font-display font-bold text-slate-900 text-[17px]">{feature.title}</h3>
                      <p className="mt-2 text-slate-500 text-sm leading-relaxed">{feature.desc}</p>
                    </div>
                  </div>
                </div>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* ---- FINAL CTA ---- */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-amethyst-600 via-amethyst-700 to-ruby-600" />
        <GridPattern />
        <FloatingOrb className="w-96 h-96 bg-white -top-40 -right-40 opacity-10" />
        <FloatingOrb className="w-72 h-72 bg-ruby-300 -bottom-20 -left-20 opacity-10" />

        <div className={`relative mx-auto max-w-4xl text-center ${isMobile ? 'px-5 py-16' : 'px-8 py-24'}`}>
          <FadeIn>
            <h2 className={`font-display font-extrabold text-white leading-tight ${isMobile ? 'text-2xl' : 'text-5xl'}`}>
              Ready to get started?
            </h2>
            <p className={`mt-5 text-amethyst-100 max-w-lg mx-auto ${isMobile ? 'text-[15px]' : 'text-lg'}`}>
              Join the platform that&apos;s changing how the Maldives shops and trades.
              Whether you&apos;re buying, selling, or exchanging — SwipeStore has you covered.
            </p>

            <div className={`mt-10 flex flex-wrap justify-center gap-4 ${isMobile ? 'flex-col' : ''}`}>
              <Link
                href="/marketplace"
                className="group inline-flex items-center justify-center gap-2 rounded-xl bg-white px-8 py-4 text-sm font-bold text-slate-900 shadow-xl hover:shadow-2xl transition-all hover:scale-[1.02] active:scale-[0.98]"
              >
                Browse Marketplace
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>
              <Link
                href="/exchange"
                className="group inline-flex items-center justify-center gap-2 rounded-xl bg-white/15 backdrop-blur border border-white/25 px-8 py-4 text-sm font-bold text-white hover:bg-white/25 transition-all hover:scale-[1.02] active:scale-[0.98]"
              >
                Trade Crypto
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>
              <Link
                href="/seller/login"
                className="group inline-flex items-center justify-center gap-2 rounded-xl bg-white/15 backdrop-blur border border-white/25 px-8 py-4 text-sm font-bold text-white hover:bg-white/25 transition-all hover:scale-[1.02] active:scale-[0.98]"
              >
                Start Selling
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>
            </div>
          </FadeIn>
        </div>
      </section>

      {/* ---- FOOTER ---- */}
      <footer className="bg-slate-900 border-t border-slate-800">
        <div className={`mx-auto max-w-7xl ${isMobile ? 'px-5 py-10' : 'px-8 py-14'}`}>
          <div className={`${isMobile ? 'flex flex-col gap-8' : 'grid grid-cols-4 gap-12'}`}>
            {/* Brand */}
            <div className={isMobile ? '' : 'col-span-2'}>
              <div className="flex items-center gap-2.5 mb-4">
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amethyst-500 to-amethyst-700 flex items-center justify-center">
                  <span className="text-white text-sm font-bold">S</span>
                </div>
                <span className="font-display font-extrabold text-white text-lg">SwipeStore</span>
              </div>
              <p className="text-sm text-slate-400 leading-relaxed max-w-sm">
                Social commerce and P2P crypto exchange for the Maldives.
                Powered by Swipe instant payments.
              </p>
            </div>

            {/* Links */}
            <div>
              <h4 className="font-display font-bold text-white text-sm mb-4">Platform</h4>
              <div className="space-y-3">
                <Link href="/marketplace" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">Marketplace</Link>
                <Link href="/exchange" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">P2P Exchange</Link>
                <Link href="/shop/island-finds-mv" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">Demo Store</Link>
              </div>
            </div>

            <div>
              <h4 className="font-display font-bold text-white text-sm mb-4">For Sellers</h4>
              <div className="space-y-3">
                <Link href="/seller/login" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">Start Selling</Link>
                <Link href="/seller/login" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">Seller Dashboard</Link>
                <Link href="/seller/login" className="block text-sm text-slate-400 hover:text-amethyst-400 transition-colors">List Crypto</Link>
              </div>
            </div>
          </div>

          {/* Bottom bar */}
          <div className={`mt-10 pt-6 border-t border-slate-800 ${isMobile ? 'text-center' : 'flex items-center justify-between'}`}>
            <p className="text-xs text-slate-500">
              &copy; 2026 SwipeStore. All rights reserved.
            </p>
            <p className={`text-xs text-slate-600 ${isMobile ? 'mt-2' : ''}`}>
              Powered by <span className="text-amethyst-400 font-semibold">Swipe</span> &middot; Built for the Maldives
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
