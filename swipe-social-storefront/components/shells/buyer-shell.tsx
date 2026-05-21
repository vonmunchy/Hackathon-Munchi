'use client'

import { useIsMobile } from '@/lib/use-device'
import { BuyerBottomBar } from '@/components/mobile/buyer-bottom-bar'
import { BuyerDesktopNav } from '@/components/desktop/buyer-desktop-nav'

interface BuyerShellProps {
  children: React.ReactNode
}

export function BuyerShell({ children }: BuyerShellProps) {
  const isMobile = useIsMobile()

  if (isMobile) {
    return (
      <div className="flex h-dvh flex-col bg-slate-50">
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
