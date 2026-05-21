'use client'

import { usePathname } from 'next/navigation'
import { useIsMobile } from '@/lib/use-device'
import { BuyerBottomBar } from '@/components/mobile/buyer-bottom-bar'
import { BuyerDesktopNav } from '@/components/desktop/buyer-desktop-nav'
import { TopBar } from '@/components/mobile/top-bar'

interface BuyerShellProps {
  children: React.ReactNode
}

function getBuyerTopBarConfig(pathname: string): { title: string; showBack: boolean } | null {
  // Top-level tabs — show back button so users can always navigate back
  if (pathname === '/marketplace') return { title: 'Marketplace', showBack: true }
  if (pathname === '/exchange') return { title: 'P2P Exchange', showBack: true }

  // Sub-pages get a back button
  if (pathname.startsWith('/exchange/buy')) return { title: 'Buy USDT', showBack: true }
  if (pathname.startsWith('/shop/') && pathname.includes('/product/')) return { title: 'Product', showBack: true }
  if (pathname.startsWith('/shop/')) return { title: 'Store', showBack: true }
  if (pathname.startsWith('/checkout/')) return { title: 'Checkout', showBack: true }
  if (pathname.startsWith('/success/')) return { title: 'Order Confirmed', showBack: false }

  return null
}

export function BuyerShell({ children }: BuyerShellProps) {
  const isMobile = useIsMobile()
  const pathname = usePathname()

  if (isMobile) {
    const topBarConfig = getBuyerTopBarConfig(pathname)

    return (
      <div className="flex h-dvh flex-col bg-slate-50">
        {topBarConfig && (
          <TopBar title={topBarConfig.title} showBack={topBarConfig.showBack} />
        )}
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
        <BuyerBottomBar />
      </div>
    )
  }

  return (
    <div className="flex h-dvh flex-col">
      <BuyerDesktopNav />
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  )
}
