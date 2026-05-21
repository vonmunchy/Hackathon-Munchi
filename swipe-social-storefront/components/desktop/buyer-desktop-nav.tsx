'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

const navLinks = [
  { label: 'Marketplace', href: '/marketplace' },
  { label: 'Exchange', href: '/exchange' },
]

export function BuyerDesktopNav() {
  const pathname = usePathname()

  function isActive(href: string) {
    return pathname.startsWith(href)
  }

  return (
    <header className="sticky top-0 z-50 shrink-0 border-b border-slate-100 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2.5">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-amethyst-600">
            <span className="font-display text-xs font-bold text-white">S</span>
          </div>
          <span className="font-display text-[15px] font-semibold tracking-tight text-slate-900">
            SwiftStore
          </span>
        </Link>

        <nav className="flex items-center gap-8">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={`text-sm font-medium transition-colors ${
                isActive(link.href)
                  ? 'text-amethyst-600'
                  : 'text-slate-500 hover:text-slate-900'
              }`}
            >
              {link.label}
            </Link>
          ))}

          <Link
            href="/seller/login"
            className="rounded-lg bg-amethyst-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amethyst-700"
          >
            Become a Seller
          </Link>
        </nav>
      </div>
    </header>
  )
}
