'use client'

import { BuyerTopNav } from '@/components/desktop/buyer-top-nav'
import { SellerSidebar } from '@/components/desktop/seller-sidebar'
import { ContentArea } from '@/components/desktop/content-area'

interface DesktopShellProps {
  isSeller: boolean
  children: React.ReactNode
}

export function DesktopShell({ isSeller, children }: DesktopShellProps) {
  if (isSeller) {
    return (
      <div className="flex h-dvh bg-slate-50">
        <SellerSidebar />
        <main className="flex-1 overflow-y-auto">
          <ContentArea>{children}</ContentArea>
        </main>
      </div>
    )
  }

  return (
    <div className="flex h-dvh flex-col">
      <BuyerTopNav />
      <main className="flex-1 overflow-y-auto">
        <ContentArea>{children}</ContentArea>
      </main>
    </div>
  )
}
