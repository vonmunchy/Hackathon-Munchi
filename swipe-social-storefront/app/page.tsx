'use client'

import Link from 'next/link'
import { useIsMobile } from '@/lib/use-device'

/* ---- Phone Mockup (CSS-only) ---- */
function PhoneMockup() {
  return (
    <div className="relative mx-auto w-[260px] h-[480px]">
      <div className="absolute inset-0 rounded-[2.5rem] border-[2px] border-slate-200 bg-white shadow-xl overflow-hidden">
        {/* Status bar */}
        <div className="h-7 bg-slate-50 flex items-center justify-between px-6">
          <span className="text-[9px] text-slate-400 font-medium">9:41</span>
          <div className="flex gap-1">
            <div className="w-3 h-1.5 rounded-sm bg-slate-300" />
            <div className="w-1.5 h-1.5 rounded-full bg-slate-300" />
          </div>
        </div>
        {/* App content */}
        <div className="px-4 pt-3">
          <div className="flex items-center gap-2 mb-4">
            <div className="w-7 h-7 rounded-full bg-amethyst-600 flex items-center justify-center">
              <span className="text-white text-[10px] font-bold">S</span>
            </div>
            <span className="text-[11px] font-semibold text-slate-800">Swipe</span>
          </div>
          {/* Exchange listing mockup */}
          <div className="mb-3 rounded-xl border border-slate-100 p-3 bg-slate-50/50">
            <div className="flex items-center justify-between mb-2">
              <span className="text-[10px] font-semibold text-slate-700">100 USDT</span>
              <span className="text-[9px] font-medium px-2 py-0.5 rounded-full bg-green-100 text-green-700">Active</span>
            </div>
            <div className="text-[9px] text-slate-500">25.50 MVR/USDT</div>
            <div className="text-[10px] font-semibold text-amethyst-600 mt-1">2,550 MVR</div>
          </div>
          {/* Product mockup */}
          <div className="mb-3 rounded-xl border border-slate-100 p-3 bg-slate-50/50">
            <div className="flex gap-2.5">
              <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-amethyst-100 to-amethyst-50 flex-shrink-0" />
              <div>
                <div className="text-[10px] font-semibold text-slate-700">Black Abaya</div>
                <div className="text-[9px] text-slate-400">Island Finds MV</div>
                <div className="text-[10px] font-semibold text-amethyst-600 mt-0.5">MVR 650</div>
              </div>
            </div>
          </div>
          {/* Bottom action */}
          <div className="mt-4 rounded-xl bg-amethyst-600 py-2.5 text-center">
            <span className="text-[10px] font-semibold text-white">Pay with Swipe</span>
          </div>
        </div>
        {/* Bottom bar */}
        <div className="absolute bottom-0 inset-x-0 h-12 border-t border-slate-100 bg-white flex items-center justify-around px-6">
          <div className="w-5 h-5 rounded bg-amethyst-100" />
          <div className="w-5 h-5 rounded bg-slate-100" />
          <div className="w-5 h-5 rounded bg-slate-100" />
        </div>
      </div>
    </div>
  )
}

/* ---- Checkmark ---- */
function CheckmarkSmall({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
    </svg>
  )
}

/* ---- Arrow ---- */
function ArrowRight() {
  return (
    <svg width="18" height="18" viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clipRule="evenodd" />
    </svg>
  )
}

/* ---- Step numbers ---- */
function StepNumber({ n }: { n: number }) {
  return (
    <div className="w-10 h-10 rounded-full bg-amethyst-100 text-amethyst-700 flex items-center justify-center text-sm font-bold flex-shrink-0">
      {n}
    </div>
  )
}

/* ======== PAGE ======== */
export default function Home() {
  const isMobile = useIsMobile()

  if (isMobile === null) {
    return <div className="min-h-screen bg-white" />
  }

  return (
    <div className="min-h-screen bg-white font-body text-slate-900">

      {/* ---- NAV BAR ---- */}
      <nav className="sticky top-0 z-50 bg-white/80 backdrop-blur-lg border-b border-slate-100">
        <div className={`mx-auto max-w-6xl flex items-center justify-between ${isMobile ? 'px-5 h-14' : 'px-8 h-16'}`}>
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-full bg-amethyst-600 flex items-center justify-center">
              <span className="text-white text-xs font-bold">S</span>
            </div>
            <span className="font-display font-bold text-slate-900 text-lg">Swipe</span>
          </div>
          {!isMobile && (
            <div className="flex items-center gap-8">
              <Link href="/marketplace" className="text-sm text-slate-500 hover:text-amethyst-600 transition-colors font-medium">Marketplace</Link>
              <Link href="/exchange" className="text-sm text-slate-500 hover:text-amethyst-600 transition-colors font-medium">Exchange</Link>
              <Link href="/seller/login" className="text-sm bg-amethyst-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-amethyst-700 transition-colors">
                Start Selling
              </Link>
            </div>
          )}
        </div>
      </nav>

      {/* ---- HERO ---- */}
      <section className={`relative overflow-hidden ${isMobile ? 'px-5 pt-12 pb-10' : 'px-8 pt-20 pb-24'}`}>
        <div className={`mx-auto max-w-6xl ${isMobile ? '' : 'flex items-center gap-16'}`}>
          {/* Text side */}
          <div className={`${isMobile ? '' : 'flex-1'}`}>
            <div className="inline-flex items-center gap-2 rounded-full bg-amethyst-50 border border-amethyst-100 px-3.5 py-1.5 mb-6">
              <div className="w-1.5 h-1.5 rounded-full bg-amethyst-500" />
              <span className="text-xs font-semibold text-amethyst-700 tracking-wide">Powered by Swipe</span>
            </div>

            <h1 className={`font-display font-extrabold tracking-tight text-slate-900 leading-[1.1] ${isMobile ? 'text-[28px]' : 'text-[44px]'}`}>
              What wasn&apos;t possible before&nbsp;&mdash; is&nbsp;now possible with&nbsp;Swipe
            </h1>

            <p className={`mt-5 text-slate-500 leading-relaxed ${isMobile ? 'text-[15px]' : 'text-lg max-w-lg'}`}>
              Shop from local sellers or buy USDT — all with instant Swipe payments. No bank transfers. No scams. No waiting.
            </p>

            {/* CTAs */}
            <div className="mt-8 flex flex-col gap-3">
              <div className={`flex gap-3 ${isMobile ? 'flex-col' : ''}`}>
                <Link
                  href="/marketplace"
                  className="inline-flex items-center justify-center gap-2 rounded-xl bg-amethyst-600 px-6 py-3 text-sm font-semibold text-white shadow-sm hover:bg-amethyst-700 transition-all hover:shadow-md"
                >
                  Shop Marketplace
                  <ArrowRight />
                </Link>
                <Link
                  href="/exchange"
                  className="inline-flex items-center justify-center gap-2 rounded-xl bg-slate-900 px-6 py-3 text-sm font-semibold text-white shadow-sm hover:bg-slate-800 transition-all hover:shadow-md"
                >
                  Buy Crypto
                  <ArrowRight />
                </Link>
              </div>
              <div className={`flex gap-4 ${isMobile ? '' : ''}`}>
                <Link href="/seller/login" className="text-sm font-medium text-slate-500 hover:text-amethyst-600 transition-colors underline decoration-slate-300 underline-offset-4 hover:decoration-amethyst-400">
                  Sell Products
                </Link>
                <Link href="/seller/login" className="text-sm font-medium text-slate-500 hover:text-amethyst-600 transition-colors underline decoration-slate-300 underline-offset-4 hover:decoration-amethyst-400">
                  Sell Crypto
                </Link>
              </div>
            </div>
          </div>

          {/* Phone mockup — desktop only */}
          {!isMobile && (
            <div className="flex-shrink-0">
              <PhoneMockup />
            </div>
          )}
        </div>
      </section>

      {/* ---- BEFORE vs AFTER ---- */}
      <section className={`bg-slate-50 ${isMobile ? 'px-5 py-12' : 'px-8 py-20'}`}>
        <div className="mx-auto max-w-6xl">
          <h2 className={`text-center font-display font-bold text-slate-900 ${isMobile ? 'text-xl' : 'text-3xl'}`}>
            Before Swipe vs. After Swipe
          </h2>
          <p className="mt-3 text-center text-slate-500 text-sm max-w-xl mx-auto">
            See how Swipe transforms the way Maldivians shop and trade
          </p>

          <div className={`mt-10 ${isMobile ? 'flex flex-col gap-8' : 'grid grid-cols-2 gap-8'}`}>
            {/* Marketplace comparison */}
            <div className="rounded-2xl border border-slate-200 bg-white overflow-hidden">
              <div className="px-6 py-4 bg-amethyst-50 border-b border-amethyst-100">
                <h3 className="font-display font-bold text-amethyst-800 text-sm">Shopping</h3>
              </div>
              <div className={`${isMobile ? '' : 'grid grid-cols-2 divide-x divide-slate-100'}`}>
                {/* Before */}
                <div className="p-5">
                  <p className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Before</p>
                  <div className="space-y-2.5">
                    {['Find seller on Instagram', 'DM to check availability', 'Get bank account details', 'Make transfer & screenshot', 'Send slip, pray it\'s verified', '~15 minutes per order'].map((item, i) => (
                      <div key={i} className="flex items-start gap-2">
                        <span className="w-1.5 h-1.5 mt-1.5 rounded-full bg-ruby-400 flex-shrink-0" />
                        <span className="text-xs text-slate-600 leading-relaxed">{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
                {/* After */}
                <div className={`p-5 ${isMobile ? 'border-t border-slate-100' : ''}`}>
                  <p className="text-xs font-bold text-amethyst-600 uppercase tracking-wider mb-3">After</p>
                  <div className="space-y-2.5">
                    {['Browse marketplace', 'Tap to buy', 'Pay with Swipe — instant', 'Order confirmed automatically', 'Under 2 minutes'].map((item, i) => (
                      <div key={i} className="flex items-start gap-2">
                        <CheckmarkSmall className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                        <span className="text-xs text-slate-700 leading-relaxed font-medium">{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Crypto comparison */}
            <div className="rounded-2xl border border-slate-200 bg-white overflow-hidden">
              <div className="px-6 py-4 bg-slate-800 border-b border-slate-700">
                <h3 className="font-display font-bold text-white text-sm">Crypto Exchange</h3>
              </div>
              <div className={`${isMobile ? '' : 'grid grid-cols-2 divide-x divide-slate-100'}`}>
                {/* Before */}
                <div className="p-5">
                  <p className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Before</p>
                  <div className="space-y-2.5">
                    {['Find seller on Telegram', 'Negotiate rate via chat', 'Transfer MVR to stranger', 'Hope USDT arrives', 'No recourse if scammed', 'High risk, no escrow'].map((item, i) => (
                      <div key={i} className="flex items-start gap-2">
                        <span className="w-1.5 h-1.5 mt-1.5 rounded-full bg-ruby-400 flex-shrink-0" />
                        <span className="text-xs text-slate-600 leading-relaxed">{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
                {/* After */}
                <div className={`p-5 ${isMobile ? 'border-t border-slate-100' : ''}`}>
                  <p className="text-xs font-bold text-amethyst-600 uppercase tracking-wider mb-3">After</p>
                  <div className="space-y-2.5">
                    {['Browse verified listings', 'See escrow proof on-chain', 'Pay with Swipe — instant', 'USDT released automatically', 'Under 60 seconds'].map((item, i) => (
                      <div key={i} className="flex items-start gap-2">
                        <CheckmarkSmall className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                        <span className="text-xs text-slate-700 leading-relaxed font-medium">{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---- HOW IT WORKS ---- */}
      <section className={`${isMobile ? 'px-5 py-12' : 'px-8 py-20'}`}>
        <div className="mx-auto max-w-4xl">
          <h2 className={`text-center font-display font-bold text-slate-900 ${isMobile ? 'text-xl' : 'text-3xl'}`}>
            How It Works
          </h2>
          <p className="mt-3 text-center text-slate-500 text-sm">
            Three steps — whether you&apos;re shopping or trading
          </p>

          <div className={`mt-12 ${isMobile ? 'flex flex-col gap-8' : 'grid grid-cols-3 gap-12'}`}>
            <div className={`flex ${isMobile ? 'items-start gap-4' : 'flex-col items-center text-center gap-4'}`}>
              <StepNumber n={1} />
              <div>
                <h3 className="font-display font-bold text-slate-900 text-base">Browse</h3>
                <p className="mt-1.5 text-slate-500 text-sm leading-relaxed">
                  Find products from local sellers or USDT listings at competitive rates
                </p>
              </div>
            </div>

            <div className={`flex ${isMobile ? 'items-start gap-4' : 'flex-col items-center text-center gap-4'}`}>
              <StepNumber n={2} />
              <div>
                <h3 className="font-display font-bold text-slate-900 text-base">Pay with Swipe</h3>
                <p className="mt-1.5 text-slate-500 text-sm leading-relaxed">
                  Instant MVR payment — no bank transfers, no screenshots, no waiting
                </p>
              </div>
            </div>

            <div className={`flex ${isMobile ? 'items-start gap-4' : 'flex-col items-center text-center gap-4'}`}>
              <StepNumber n={3} />
              <div>
                <h3 className="font-display font-bold text-slate-900 text-base">Done</h3>
                <p className="mt-1.5 text-slate-500 text-sm leading-relaxed">
                  Products shipped to your door. USDT transferred to your wallet. Instant confirmation.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---- WHY SWIPE ---- */}
      <section className={`bg-slate-50 ${isMobile ? 'px-5 py-12' : 'px-8 py-20'}`}>
        <div className="mx-auto max-w-5xl">
          <h2 className={`text-center font-display font-bold text-slate-900 ${isMobile ? 'text-xl' : 'text-3xl'}`}>
            Why Swipe
          </h2>
          <p className="mt-3 text-center text-slate-500 text-sm max-w-md mx-auto">
            Built for how the Maldives actually does business
          </p>

          <div className={`mt-10 ${isMobile ? 'flex flex-col gap-4' : 'grid grid-cols-2 gap-5'}`}>
            {/* No Fake Transfer Slips */}
            <div className="rounded-xl border border-slate-200 bg-white p-6 hover:border-amethyst-200 hover:shadow-sm transition-all">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-ruby-50 flex items-center justify-center flex-shrink-0">
                  <svg className="w-5 h-5 text-ruby-500" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-bold text-slate-900 text-[15px]">No Scams</h3>
                  <p className="mt-1 text-slate-500 text-sm leading-relaxed">
                    Payment is verified by Swipe before anything ships or transfers. No fake screenshots, no chargebacks.
                  </p>
                </div>
              </div>
            </div>

            {/* Instant Reconciliation */}
            <div className="rounded-xl border border-slate-200 bg-white p-6 hover:border-amethyst-200 hover:shadow-sm transition-all">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-amethyst-50 flex items-center justify-center flex-shrink-0">
                  <svg className="w-5 h-5 text-amethyst-600" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-bold text-slate-900 text-[15px]">Instant Settlement</h3>
                  <p className="mt-1 text-slate-500 text-sm leading-relaxed">
                    Swipe confirms payment in seconds. Sellers see revenue instantly. Crypto releases automatically.
                  </p>
                </div>
              </div>
            </div>

            {/* Transparent */}
            <div className="rounded-xl border border-slate-200 bg-white p-6 hover:border-amethyst-200 hover:shadow-sm transition-all">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-green-50 flex items-center justify-center flex-shrink-0">
                  <svg className="w-5 h-5 text-green-600" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                    <path fillRule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-bold text-slate-900 text-[15px]">Fully Transparent</h3>
                  <p className="mt-1 text-slate-500 text-sm leading-relaxed">
                    Track every order and every trade. Escrow balances visible in real-time. Nothing hidden.
                  </p>
                </div>
              </div>
            </div>

            {/* Built for Maldives */}
            <div className="rounded-xl border border-slate-200 bg-white p-6 hover:border-amethyst-200 hover:shadow-sm transition-all">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-blue-50 flex items-center justify-center flex-shrink-0">
                  <svg className="w-5 h-5 text-blue-600" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-bold text-slate-900 text-[15px]">Built for Maldives</h3>
                  <p className="mt-1 text-slate-500 text-sm leading-relaxed">
                    MVR payments, Maldivian sellers, local delivery. Designed for how we actually do business here.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---- FINAL CTA ---- */}
      <section className={`${isMobile ? 'px-5 py-12' : 'px-8 py-20'}`}>
        <div className="mx-auto max-w-3xl text-center">
          <h2 className={`font-display font-bold text-slate-900 ${isMobile ? 'text-xl' : 'text-3xl'}`}>
            Ready to get started?
          </h2>
          <p className="mt-3 text-slate-500 text-sm max-w-md mx-auto">
            Whether you&apos;re buying, selling, or trading — Swipe makes it instant.
          </p>

          <div className={`mt-8 flex flex-wrap justify-center gap-3 ${isMobile ? 'flex-col' : ''}`}>
            <Link
              href="/marketplace"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-amethyst-600 px-6 py-3 text-sm font-semibold text-white hover:bg-amethyst-700 transition-colors"
            >
              Shop Marketplace
            </Link>
            <Link
              href="/exchange"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-slate-900 px-6 py-3 text-sm font-semibold text-white hover:bg-slate-800 transition-colors"
            >
              Buy Crypto
            </Link>
            <Link
              href="/seller/login"
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-200 px-6 py-3 text-sm font-semibold text-slate-700 hover:border-amethyst-300 hover:text-amethyst-700 transition-colors"
            >
              Sell Products
            </Link>
            <Link
              href="/seller/login"
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-200 px-6 py-3 text-sm font-semibold text-slate-700 hover:border-amethyst-300 hover:text-amethyst-700 transition-colors"
            >
              Sell Crypto
            </Link>
          </div>
        </div>
      </section>

      {/* ---- FOOTER ---- */}
      <footer className="border-t border-slate-100 bg-slate-50">
        <div className={`mx-auto max-w-6xl ${isMobile ? 'px-5 py-8' : 'px-8 py-10'}`}>
          <div className={`flex ${isMobile ? 'flex-col items-center gap-5' : 'items-center justify-between'}`}>
            <div className="flex items-center gap-2.5">
              <div className="w-7 h-7 rounded-full bg-amethyst-600 flex items-center justify-center">
                <span className="text-white text-[10px] font-bold">S</span>
              </div>
              <span className="font-display font-bold text-slate-800 text-sm">Swipe</span>
            </div>
            <div className={`flex flex-wrap gap-6 text-sm ${isMobile ? 'justify-center' : ''}`}>
              <Link href="/marketplace" className="text-slate-500 hover:text-amethyst-600 transition-colors">Marketplace</Link>
              <Link href="/exchange" className="text-slate-500 hover:text-amethyst-600 transition-colors">Exchange</Link>
              <Link href="/shop/island-finds-mv" className="text-slate-500 hover:text-amethyst-600 transition-colors">Demo Store</Link>
              <Link href="/seller/login" className="text-slate-500 hover:text-amethyst-600 transition-colors">Seller Login</Link>
            </div>
          </div>
          <p className={`mt-6 text-xs text-slate-400 ${isMobile ? 'text-center' : ''}`}>
            &copy; 2026 Swipe. All rights reserved. Built for the Maldives.
          </p>
        </div>
      </footer>
    </div>
  )
}
