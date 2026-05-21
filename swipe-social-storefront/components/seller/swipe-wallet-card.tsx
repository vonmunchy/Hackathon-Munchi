'use client'

import { formatMVR } from '@/lib/format'

interface SwipeWalletCardProps {
  available: number
  pending: number
  loading: boolean
  error?: string | null
  onRefresh: () => void
}

export function SwipeWalletCard({ available, pending, loading, error, onRefresh }: SwipeWalletCardProps) {
  return (
    <div className="bg-gradient-to-r from-amethyst-600 to-amethyst-500 rounded-lg p-5 text-white relative overflow-hidden">
      {/* Decorative circle */}
      <div className="absolute -right-6 -top-6 w-24 h-24 rounded-full bg-white/10" />

      <div className="flex items-start justify-between">
        <div>
          <p className="text-amethyst-100 text-sm font-medium">Available Balance</p>
          <p className="text-2xl font-mono font-bold mt-1">
            {loading ? '---' : error ? 'Error' : formatMVR(available)}
          </p>
          <p className="text-amethyst-200 text-sm mt-2">
            Pending: {loading ? '---' : formatMVR(pending)}
          </p>
        </div>
        <button
          onClick={onRefresh}
          disabled={loading}
          className="p-2 rounded-full hover:bg-white/10 transition-colors disabled:opacity-50"
          title="Refresh balance"
        >
          <svg
            className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        </button>
      </div>

      {error && (
        <p className="text-ruby-200 text-xs mt-2">{error}</p>
      )}
    </div>
  )
}
