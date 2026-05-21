'use client'

import { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { SESSION_COOKIE_NAME } from '@/lib/constants'

export default function SellerLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const pathname = usePathname()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    // Skip auth check for login page
    if (pathname === '/seller/login') {
      setChecked(true)
      return
    }

    // Check if cookie exists (client-side check only)
    const hasCookie = document.cookie
      .split(';')
      .some((c) => c.trim().startsWith('seller_logged_in='))

    if (!hasCookie) {
      router.replace('/seller/login')
    } else {
      // Sync store info from cookies to localStorage (set by OAuth callback)
      const cookies = Object.fromEntries(
        document.cookie.split(';').map((c) => {
          const [k, ...v] = c.trim().split('=')
          return [k, decodeURIComponent(v.join('='))]
        }),
      )
      if (cookies.store_slug) {
        localStorage.setItem('storeSlug', cookies.store_slug)
      }
      if (cookies.store_name) {
        localStorage.setItem('storeName', cookies.store_name)
      }

      // Bootstrap Convex session if seller_session_token cookie is missing
      const hasSessionToken = document.cookie
        .split(';')
        .some((c) => c.trim().startsWith('seller_session_token='))
      if (!hasSessionToken) {
        fetch('/api/auth/session').then((res) => {
          if (res.ok) {
            // Reload to pick up the new cookie
            window.location.reload()
          } else {
            // Session expired or lost - re-login required
            router.replace('/seller/login')
          }
        }).catch(() => {
          router.replace('/seller/login')
        })
        return
      }

      // Redirect to onboarding if not completed (skip if already on onboarding page)
      const onboardingDone = document.cookie
        .split(';')
        .some((c) => c.trim().startsWith('onboarding_complete='))
      if (!onboardingDone && pathname !== '/seller/onboarding') {
        router.replace('/seller/onboarding')
        return
      }

      setChecked(true)
    }
  }, [pathname, router])

  // Don't render children until auth check completes (unless on login page)
  if (!checked && pathname !== '/seller/login') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-amethyst-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  return <>{children}</>
}
