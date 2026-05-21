'use client'

import Link from 'next/link'
import { useIsMobile } from '@/lib/use-device'

/* ---- Inline SVG Icons ---- */
function ShoppingBagIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M14 16V12a10 10 0 0120 0v4" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
      <rect x="8" y="16" width="32" height="26" rx="4" stroke="currentColor" strokeWidth="2.5" />
      <circle cx="18" cy="24" r="2" fill="currentColor" opacity="0.4" />
      <circle cx="30" cy="24" r="2" fill="currentColor" opacity="0.4" />
    </svg>
  )
}

function ShieldIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M24 4L8 12v12c0 11 16 18 16 18s16-7 16-18V12L24 4z" stroke="currentColor" strokeWidth="2.5" strokeLinejoin="round" />
      <path d="M18 24l4 4 8-8" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function BrowseIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="20" cy="20" r="14" stroke="currentColor" strokeWidth="2.5" />
      <path d="M30 30l12 12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
      <rect x="14" y="14" width="12" height="4" rx="2" fill="currentColor" opacity="0.3" />
      <rect x="14" y="22" width="8" height="4" rx="2" fill="currentColor" opacity="0.2" />
    </svg>
  )
}

function PaymentIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="4" y="10" width="40" height="28" rx="6" stroke="currentColor" strokeWidth="2.5" />
      <path d="M4 20h40" stroke="currentColor" strokeWidth="2.5" />
      <circle cx="34" cy="30" r="4" fill="currentColor" opacity="0.3" />
      <circle cx="28" cy="30" r="4" fill="currentColor" opacity="0.2" />
    </svg>
  )
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="24" cy="24" r="20" stroke="currentColor" strokeWidth="2.5" />
      <path d="M14 24l7 7 13-13" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function NoScamIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="24" cy="24" r="20" stroke="currentColor" strokeWidth="2.5" />
      <path d="M16 16l16 16M32 16L16 32" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
    </svg>
  )
}

function InstantIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M26 4L12 28h12l-2 16 14-24H24l2-16z" stroke="currentColor" strokeWidth="2.5" strokeLinejoin="round" fill="currentColor" fillOpacity="0.15" />
    </svg>
  )
}

function TransparentIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="24" cy="24" r="20" stroke="currentColor" strokeWidth="2.5" />
      <circle cx="24" cy="24" r="8" stroke="currentColor" strokeWidth="2" fill="currentColor" fillOpacity="0.15" />
      <circle cx="24" cy="24" r="2" fill="currentColor" />
    </svg>
  )
}

function MaldivesIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="6" y="12" width="36" height="24" rx="4" stroke="currentColor" strokeWidth="2.5" />
      <circle cx="24" cy="24" r="6" stroke="currentColor" strokeWidth="2" fill="currentColor" fillOpacity="0.15" />
    </svg>
  )
}

/* ---- Bullet point component ---- */
function BulletPoint({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-start gap-3">
      <span className="mt-1.5 flex-shrink-0 w-2 h-2 rounded-full bg-amethyst-500" />
      <span className="text-slate-300 text-sm leading-relaxed">{children}</span>
    </li>
  )
}

/* ======== PAGE ======== */
export default function Home() {
  const isMobile = useIsMobile()

  if (isMobile === null) {
    return <div className="min-h-screen bg-slate-900" />
  }

  return (
    <div className="min-h-screen bg-slate-900 font-body text-white">
      {/* ---- HERO ---- */}
      <section className={`relative overflow-hidden ${isMobile ? 'px-6 pt-16 pb-14' : 'px-8 pt-28 pb-24'}`}>
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-900 to-amethyst-900/10" />

        <div className="relative mx-auto max-w-6xl text-center">
          <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-amethyst-400 mb-5">
            Powered by Swipe
          </p>
          <h1 className={`font-display font-extrabold tracking-[-0.03em] text-white leading-[1.08] ${isMobile ? 'text-3xl' : 'text-[52px]'}`}>
            What wasn&apos;t possible before &mdash; is now possible with Swipe
          </h1>
          <p className={`mt-5 text-slate-400 leading-relaxed mx-auto ${isMobile ? 'text-base max-w-sm' : 'text-lg max-w-2xl'}`}>
            Instant payments. Zero trust required. Shop and trade in the Maldives like never before.
          </p>

          {/* CTAs */}
          <div className={`mt-10 flex flex-col gap-4 ${isMobile ? '' : 'items-center'}`}>
            {/* Buyer row */}
            <div className={`flex gap-3 ${isMobile ? 'flex-col' : 'flex-row justify-center'}`}>
              <Link
                href="/marketplace"
                className="inline-flex items-center justify-center gap-2 rounded-xl bg-amethyst-600 px-7 py-3.5 text-sm font-semibold text-white shadow-md transition-all hover:bg-amethyst-700 hover:shadow-lg"
              >
                Shop Marketplace
              </Link>
              <Link
                href="/exchange"
                className="inline-flex items-center justify-center gap-2 rounded-xl bg-amethyst-600 px-7 py-3.5 text-sm font-semibold text-white shadow-md transition-all hover:bg-amethyst-700 hover:shadow-lg"
              >
                Buy Crypto
              </Link>
            </div>
            {/* Seller row */}
            <div className={`flex gap-3 ${isMobile ? 'flex-col' : 'flex-row justify-center'}`}>
              <Link
                href="/seller/login"
                className="inline-flex items-center justify-center gap-2 rounded-xl border border-amethyst-500 px-7 py-3.5 text-sm font-semibold text-amethyst-300 transition-all hover:bg-amethyst-600/10 hover:text-amethyst-200"
              >
                Sell Products
              </Link>
              <Link
                href="/seller/login"
                className="inline-flex items-center justify-center gap-2 rounded-xl border border-amethyst-500 px-7 py-3.5 text-sm font-semibold text-amethyst-300 transition-all hover:bg-amethyst-600/10 hover:text-amethyst-200"
              >
                Sell Crypto
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ---- TWO PILLARS ---- */}
      <section className={`${isMobile ? 'px-6 py-14' : 'px-8 py-20'}`}>
        <div className={`mx-auto max-w-6xl ${isMobile ? 'flex flex-col gap-6' : 'grid grid-cols-2 gap-8'}`}>
          {/* Marketplace Card */}
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-8">
            <ShoppingBagIcon className="w-12 h-12 text-amethyst-400 mb-5" />
            <h3 className={`font-display font-bold text-white ${isMobile ? 'text-xl' : 'text-2xl'}`}>
              Social Commerce, Simplified
            </h3>
            <ul className="mt-5 space-y-3">
              <BulletPoint>Browse products from local sellers</BulletPoint>
              <BulletPoint>Pay instantly with Swipe — no bank transfers</BulletPoint>
              <BulletPoint>No more fake transfer slips</BulletPoint>
            </ul>
            <Link
              href="/marketplace"
              className="mt-6 inline-flex items-center gap-1.5 text-amethyst-400 font-semibold text-sm hover:text-amethyst-300 transition-colors"
            >
              Browse Marketplace
              <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            </Link>
          </div>

          {/* Crypto Exchange Card */}
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-8">
            <ShieldIcon className="w-12 h-12 text-amethyst-400 mb-5" />
            <h3 className={`font-display font-bold text-white ${isMobile ? 'text-xl' : 'text-2xl'}`}>
              P2P USDT Trading, Secured
            </h3>
            <ul className="mt-5 space-y-3">
              <BulletPoint>Buy USDT at competitive rates</BulletPoint>
              <BulletPoint>Platform escrow protects every trade</BulletPoint>
              <BulletPoint>Instant release on payment confirmation</BulletPoint>
            </ul>
            <Link
              href="/exchange"
              className="mt-6 inline-flex items-center gap-1.5 text-amethyst-400 font-semibold text-sm hover:text-amethyst-300 transition-colors"
            >
              Buy USDT
              <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            </Link>
          </div>
        </div>
      </section>

      {/* ---- HOW IT WORKS ---- */}
      <section className={`${isMobile ? 'px-6 py-14' : 'px-8 py-20'} border-t border-slate-800`}>
        <div className="mx-auto max-w-5xl">
          <h2 className={`text-center font-display font-bold text-white ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
            How It Works
          </h2>
          <p className="mt-3 text-center text-slate-400 text-base">
            Three steps to get what you need
          </p>

          <div className={`mt-12 ${isMobile ? 'flex flex-col gap-10' : 'grid grid-cols-3 gap-10'}`}>
            {/* Step 1 */}
            <div className="flex flex-col items-center text-center">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-amethyst-600/20 text-amethyst-400 font-bold text-sm mb-4">
                1
              </div>
              <BrowseIcon className="w-14 h-14 text-amethyst-400 mb-4" />
              <h3 className="text-lg font-bold text-white">Browse</h3>
              <p className="mt-2 text-slate-400 text-sm leading-relaxed max-w-xs">
                Find what you need — products or USDT listings
              </p>
            </div>

            {/* Step 2 */}
            <div className="flex flex-col items-center text-center">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-amethyst-600/20 text-amethyst-400 font-bold text-sm mb-4">
                2
              </div>
              <PaymentIcon className="w-14 h-14 text-amethyst-400 mb-4" />
              <h3 className="text-lg font-bold text-white">Pay with Swipe</h3>
              <p className="mt-2 text-slate-400 text-sm leading-relaxed max-w-xs">
                Instant MVR payment, no bank transfers needed
              </p>
            </div>

            {/* Step 3 */}
            <div className="flex flex-col items-center text-center">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-amethyst-600/20 text-amethyst-400 font-bold text-sm mb-4">
                3
              </div>
              <CheckIcon className="w-14 h-14 text-amethyst-400 mb-4" />
              <h3 className="text-lg font-bold text-white">Done</h3>
              <p className="mt-2 text-slate-400 text-sm leading-relaxed max-w-xs">
                Products shipped or USDT transferred in seconds
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* ---- TRUST SECTION ---- */}
      <section className={`${isMobile ? 'px-6 py-14' : 'px-8 py-20'} border-t border-slate-800`}>
        <div className="mx-auto max-w-5xl">
          <h2 className={`text-center font-display font-bold text-white ${isMobile ? 'text-2xl' : 'text-4xl'}`}>
            Built for the Maldives
          </h2>

          <div className={`mt-10 ${isMobile ? 'flex flex-col gap-5' : 'grid grid-cols-2 gap-6'}`}>
            {/* No Scams */}
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6 flex items-start gap-4">
              <NoScamIcon className="w-10 h-10 text-amethyst-400 flex-shrink-0" />
              <div>
                <h4 className="font-bold text-white text-base">No Scams</h4>
                <p className="mt-1 text-slate-400 text-sm leading-relaxed">
                  Escrow and instant verification eliminate fraud
                </p>
              </div>
            </div>

            {/* Instant Settlement */}
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6 flex items-start gap-4">
              <InstantIcon className="w-10 h-10 text-amethyst-400 flex-shrink-0" />
              <div>
                <h4 className="font-bold text-white text-base">Instant Settlement</h4>
                <p className="mt-1 text-slate-400 text-sm leading-relaxed">
                  No waiting for bank confirmations
                </p>
              </div>
            </div>

            {/* Transparent */}
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6 flex items-start gap-4">
              <TransparentIcon className="w-10 h-10 text-amethyst-400 flex-shrink-0" />
              <div>
                <h4 className="font-bold text-white text-base">Transparent</h4>
                <p className="mt-1 text-slate-400 text-sm leading-relaxed">
                  Track every transaction in real-time
                </p>
              </div>
            </div>

            {/* Maldivian First */}
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6 flex items-start gap-4">
              <MaldivesIcon className="w-10 h-10 text-amethyst-400 flex-shrink-0" />
              <div>
                <h4 className="font-bold text-white text-base">Maldivian First</h4>
                <p className="mt-1 text-slate-400 text-sm leading-relaxed">
                  Built for how we actually do business here
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---- FOOTER ---- */}
      <footer className="border-t border-slate-800 px-6 py-10">
        <div className="mx-auto max-w-5xl">
          <div className={`flex ${isMobile ? 'flex-col items-center gap-4' : 'items-center justify-between'}`}>
            <div className={`flex flex-wrap gap-5 text-sm ${isMobile ? 'justify-center' : ''}`}>
              <Link href="/marketplace" className="text-slate-400 hover:text-amethyst-400 transition-colors">
                Marketplace
              </Link>
              <Link href="/exchange" className="text-slate-400 hover:text-amethyst-400 transition-colors">
                Exchange
              </Link>
              <Link href="/seller/login" className="text-slate-400 hover:text-amethyst-400 transition-colors">
                Sell Products
              </Link>
              <Link href="/seller/login" className="text-slate-400 hover:text-amethyst-400 transition-colors">
                Sell Crypto
              </Link>
            </div>
            <p className={`text-sm font-medium text-amethyst-400 ${isMobile ? 'mt-2' : ''}`}>
              Powered by Swipe
            </p>
          </div>
          <p className={`mt-6 text-slate-500 text-xs ${isMobile ? 'text-center' : ''}`}>
            &copy; 2026 Swipe. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  )
}
