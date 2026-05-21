'use client'

import Link from 'next/link'

export function BuyerTopNav() {
  return (
    <header className="shrink-0 border-b border-slate-100 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between px-6">
        <Link
          href="/shop/island-finds-mv"
          className="flex items-center gap-2.5"
        >
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-amethyst-600">
            <span className="font-display text-xs font-bold text-white">S</span>
          </div>
          <span className="font-display text-[15px] font-semibold tracking-tight text-slate-900">
            SwiftStore
          </span>
        </Link>
        <div />
      </div>
    </header>
  )
}
