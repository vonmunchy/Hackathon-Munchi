'use client'

import { usePathname } from 'next/navigation'
import { useIsMobile } from '@/lib/use-device'
import { SplashLoader } from '@/components/shells/splash-loader'
import { MobileShell } from '@/components/shells/mobile-shell'
import { DesktopShell } from '@/components/shells/desktop-shell'
import { BuyerShell } from '@/components/shells/buyer-shell'

export function LayoutRouter({ children }: { children: React.ReactNode }) {
  const isMobile = useIsMobile()
  const pathname = usePathname()

  const isSeller = pathname.startsWith('/seller') && pathname !== '/seller/login'
  const isBuyer = pathname.startsWith('/marketplace') || pathname.startsWith('/exchange')
  const isLanding = pathname === '/'
  const isBackstage = pathname.startsWith('/backstage')

  if (isMobile === null) {
    return <SplashLoader />
  }

  // Landing page and backstage render without any shell
  if (isLanding || isBackstage) {
    return <>{children}</>
  }

  if (isBuyer) {
    return <BuyerShell>{children}</BuyerShell>
  }

  if (isMobile) {
    return (
      <MobileShell isSeller={isSeller}>
        {children}
      </MobileShell>
    )
  }

  return (
    <DesktopShell isSeller={isSeller}>
      {children}
    </DesktopShell>
  )
}
