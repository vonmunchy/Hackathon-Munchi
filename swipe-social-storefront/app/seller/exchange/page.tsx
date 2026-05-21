'use client'

import { useState } from 'react'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import Link from 'next/link'
import type { Id } from '@/convex/_generated/dataModel'

const statusColors: Record<string, string> = {
  active: 'bg-green-100 text-green-700',
  pending_deposit: 'bg-yellow-100 text-yellow-700',
  withdrawn: 'bg-slate-200 text-slate-600',
  sold_out: 'bg-blue-100 text-blue-700',
}

function truncateAddress(addr: string) {
  if (addr.length <= 10) return addr
  return addr.slice(0, 4) + '...' + addr.slice(-6)
}

export default function ExchangeDashboardPage() {
  const sessionToken = useSessionToken()

  const listings = useQuery(
    api.exchange.getSellerListings,
    sessionToken ? { sessionToken } : 'skip',
  )

  const withdrawListing = useMutation(api.exchange.withdrawListing)

  const [withdrawingId, setWithdrawingId] = useState<string | null>(null)
  const [confirmWithdraw, setConfirmWithdraw] = useState<{ id: string; amount: number; wallet: string } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [copiedAddr, setCopiedAddr] = useState<string | null>(null)

  const activeListings = listings?.filter(l => l.status === 'active') ?? []
  const totalEscrowed = activeListings.reduce((sum, l) => sum + l.availableBalance + l.reservedBalance, 0)
  const activeCount = activeListings.length

  async function handleWithdraw() {
    if (!confirmWithdraw || !sessionToken) return
    setWithdrawingId(confirmWithdraw.id)
    setError(null)
    try {
      await withdrawListing({
        sessionToken,
        listingId: confirmWithdraw.id as Id<'exchangeListings'>,
      })
      setSuccess(`Successfully withdrew ${confirmWithdraw.amount} USDT`)
      setConfirmWithdraw(null)
      setTimeout(() => setSuccess(null), 4000)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Withdrawal failed'
      if (message.includes('active reservations')) {
        setError('Cannot withdraw — pending purchases in progress')
      } else {
        setError(message)
      }
      setConfirmWithdraw(null)
    } finally {
      setWithdrawingId(null)
    }
  }

  async function copyAddress(addr: string) {
    try {
      await navigator.clipboard.writeText(addr)
      setCopiedAddr(addr)
      setTimeout(() => setCopiedAddr(null), 2000)
    } catch { /* ignore */ }
  }

  return (
    <div className="p-4 md:p-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-display font-bold text-slate-900">Crypto Exchange</h1>
        <Link
          href="/seller/exchange/new"
          className="inline-flex items-center gap-1.5 bg-amethyst-500 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-amethyst-600 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create Listing
        </Link>
      </div>

      {/* Success / Error messages */}
      {success && (
        <div className="bg-green-50 text-green-700 text-sm rounded-lg p-3 mb-4">
          {success}
        </div>
      )}
      {error && (
        <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">
          {error}
        </div>
      )}

      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">Total USDT Escrowed</p>
          <p className="text-2xl font-mono font-bold text-slate-900 mt-1">{totalEscrowed.toFixed(2)}</p>
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">Active Listings</p>
          <p className="text-2xl font-mono font-bold text-slate-900 mt-1">{activeCount}</p>
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">Completed Trades</p>
          <p className="text-2xl font-mono font-bold text-slate-900 mt-1">0</p>
        </div>
      </div>

      {/* Listings */}
      {listings === undefined ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-100 rounded-xl h-28 animate-pulse" />
          ))}
        </div>
      ) : listings.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <svg className="w-12 h-12 mx-auto mb-3 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-lg font-medium">No listings yet</p>
          <p className="text-sm mt-1">Create your first USDT listing to start selling.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {listings.map((listing) => (
            <div key={listing._id} className="bg-white rounded-xl border border-slate-200 p-4 space-y-3">
              {/* Status + Amount */}
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-lg font-mono font-bold text-slate-900">{listing.usdtAmount} USDT</p>
                  <p className="text-sm text-slate-500">Rate: {listing.rate} MVR/USDT</p>
                </div>
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[listing.status] ?? 'bg-slate-100 text-slate-500'}`}>
                  {listing.status.replace('_', ' ')}
                </span>
              </div>

              {/* Balances */}
              <div className="flex gap-4 text-sm">
                <div>
                  <span className="text-slate-500">Available:</span>{' '}
                  <span className="font-mono font-medium text-slate-900">{listing.availableBalance}</span>
                </div>
                <div>
                  <span className="text-slate-500">Reserved:</span>{' '}
                  <span className="font-mono font-medium text-slate-900">{listing.reservedBalance}</span>
                </div>
              </div>

              {/* Escrow wallet */}
              <div className="flex items-center gap-2">
                <p className="text-xs text-slate-400">Escrow:</p>
                <code className="text-xs font-mono text-slate-600 bg-slate-50 px-2 py-0.5 rounded">
                  {truncateAddress(listing.walletAddress)}
                </code>
                <button
                  onClick={() => copyAddress(listing.walletAddress)}
                  className="text-slate-400 hover:text-amethyst-600 transition-colors"
                  title="Copy address"
                >
                  {copiedAddr === listing.walletAddress ? (
                    <svg className="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                  )}
                </button>
              </div>

              {/* Actions */}
              {listing.status === 'active' && listing.reservedBalance === 0 && (
                <button
                  onClick={() => setConfirmWithdraw({ id: listing._id, amount: listing.availableBalance, wallet: listing.walletAddress })}
                  disabled={withdrawingId === listing._id}
                  className="w-full mt-2 bg-slate-100 text-slate-700 py-2 rounded-lg text-sm font-medium hover:bg-slate-200 transition-colors disabled:opacity-50"
                >
                  {withdrawingId === listing._id ? 'Withdrawing...' : 'Withdraw'}
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Withdraw confirmation modal */}
      {confirmWithdraw && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-sm w-full p-6 space-y-4">
            <h3 className="text-lg font-display font-bold text-slate-900">Confirm Withdrawal</h3>
            <p className="text-sm text-slate-600">
              Withdraw <span className="font-mono font-semibold">{confirmWithdraw.amount} USDT</span> to{' '}
              <code className="text-xs bg-slate-100 px-1 py-0.5 rounded font-mono">{truncateAddress(confirmWithdraw.wallet)}</code>?
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setConfirmWithdraw(null)}
                className="flex-1 py-2 rounded-lg border border-slate-300 text-sm font-medium text-slate-600 hover:bg-slate-50"
              >
                Cancel
              </button>
              <button
                onClick={handleWithdraw}
                className="flex-1 py-2 rounded-lg bg-amethyst-500 text-white text-sm font-medium hover:bg-amethyst-600"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
