'use client'

import { TopBar } from '@/components/mobile/top-bar'
import { BottomTabBar } from '@/components/mobile/bottom-tab-bar'

interface MobileShellProps {
  isSeller: boolean
  children: React.ReactNode
}

export function MobileShell({ isSeller, children }: MobileShellProps) {
  return (
    <div className="flex h-dvh flex-col bg-slate-50">
      <TopBar title={isSeller ? 'Seller Dashboard' : 'SwiftStore'} showBack={isSeller} />
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
      {isSeller && <BottomTabBar />}
    </div>
  )
}
