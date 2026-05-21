'use client'

import { Suspense, useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'

function LoginContent() {
  const searchParams = useSearchParams()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const err = searchParams.get('error')
    if (err) {
      setError(err)
      window.history.replaceState({}, '', '/seller/login')
    }
  }, [searchParams])

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 px-4">
      <div className="w-full max-w-sm">
        {/* Logo/Brand */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-amethyst-600 text-white mb-4">
            <span className="font-display text-xl font-bold">S</span>
          </div>
          <h1 className="text-2xl font-display font-bold tracking-tight text-slate-900">Seller Portal</h1>
          <p className="text-slate-500 text-sm mt-1">Sign in to manage your store</p>
        </div>

        {/* Login Options */}
        <div className="bg-white rounded-xl border border-slate-100 shadow-md p-6 space-y-4">
          {error && (
            <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3">
              {error}
            </div>
          )}

          {/* Meta Business Login */}
          <a
            href="/api/auth/meta/facebook"
            className="w-full flex items-center justify-center gap-3 bg-[#0082FB] text-white py-3.5 rounded-lg font-medium hover:bg-[#006FDB] transition-colors"
          >
            <svg className="w-5 h-5" viewBox="0 0 36 36" fill="currentColor">
              <path d="M18 0C8.059 0 0 8.059 0 18s8.059 18 18 18 18-8.059 18-18S27.941 0 18 0zm7.902 11.2c-.732 1.092-2.736 4.308-4.392 6.876l-.096.144c-.36.54-.636.96-.816 1.14-.264.276-.552.276-.852-.06-.24-.264-.624-.852-1.2-1.68-.576-.828-1.14-1.692-1.86-2.676-1.284-1.764-2.472-2.964-3.792-2.964-1.488 0-2.88 1.896-4.188 4.596-.756 1.572-1.548 3.444-2.04 4.776-.696 1.908-.924 2.628-.924 2.628s-.12-.372.12-1.5c.168-.768.684-2.508 1.32-4.128.852-2.16 1.86-4.38 2.664-5.496C11.454 11.028 13.11 9.6 14.85 9.6c1.992 0 3.756 1.512 5.436 4.272.564.924 1.08 1.86 1.596 2.76.12.216.24.42.36.624.36-.576.684-1.104.924-1.464 1.368-2.1 3.324-4.656 5.388-6.192.36 0 .54.348-.072 1.092-.204.252-.384.468-.576.696v-.192l.012.012c-.708.396-1.332 1.02-2.016 1.992z" />
            </svg>
            Continue with Meta Business
          </a>

          <p className="text-center text-xs text-slate-400 pt-2">
            Connect your Facebook Page and Instagram to manage your store
          </p>
        </div>
      </div>
    </div>
  )
}

export default function SellerLoginPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-amethyst-500 border-t-transparent rounded-full animate-spin" />
      </div>
    }>
      <LoginContent />
    </Suspense>
  )
}
