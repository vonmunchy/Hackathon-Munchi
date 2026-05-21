'use client'

import { useState } from 'react'
import { useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import type { Id } from '@/convex/_generated/dataModel'

export default function NewExchangeListingPage() {
  const sessionToken = useSessionToken()
  const router = useRouter()

  const createListing = useMutation(api.exchange.createListing)
  const confirmDeposit = useMutation(api.exchange.confirmDeposit)

  // Step 1 state
  const [usdtAmount, setUsdtAmount] = useState('')
  const [rate, setRate] = useState('')
  const [partialAllowed, setPartialAllowed] = useState(true)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Step 2 state
  const [listingId, setListingId] = useState<string | null>(null)
  const [walletAddress, setWalletAddress] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const [deposited, setDeposited] = useState(false)
  const [copied, setCopied] = useState(false)

  const parsedAmount = parseFloat(usdtAmount)
  const parsedRate = parseFloat(rate)
  const calculatedMVR = !isNaN(parsedAmount) && !isNaN(parsedRate) ? parsedAmount * parsedRate : 0

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!sessionToken) return
    setError(null)

    if (isNaN(parsedAmount) || parsedAmount < 1) {
      setError('USDT amount must be at least 1')
      return
    }
    if (isNaN(parsedRate) || parsedRate <= 0) {
      setError('Rate must be greater than 0')
      return
    }

    setCreating(true)
    try {
      const result = await createListing({
        sessionToken,
        usdtAmount: parsedAmount,
        rate: parsedRate,
        partialAllowed,
      })
      setListingId(result.listingId)
      setWalletAddress(result.walletAddress)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create listing')
    } finally {
      setCreating(false)
    }
  }

  async function handleConfirmDeposit() {
    if (!sessionToken || !listingId) return
    setConfirming(true)
    setError(null)
    try {
      await confirmDeposit({
        sessionToken,
        listingId: listingId as Id<'exchangeListings'>,
      })
      setDeposited(true)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to confirm deposit')
    } finally {
      setConfirming(false)
    }
  }

  async function copyWallet() {
    if (!walletAddress) return
    try {
      await navigator.clipboard.writeText(walletAddress)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
  }

  // Step 3: Success
  if (deposited) {
    return (
      <div className="p-4 md:p-6 max-w-lg mx-auto">
        <div className="text-center py-12 space-y-4">
          <div className="w-16 h-16 mx-auto rounded-full bg-green-100 flex items-center justify-center">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-xl font-display font-bold text-slate-900">Listing is now live!</h2>
          <p className="text-sm text-slate-500">
            Your {parsedAmount} USDT listing at {parsedRate} MVR/USDT is active and visible to buyers.
          </p>
          <Link
            href="/seller/exchange"
            className="inline-flex items-center gap-1.5 bg-amethyst-500 text-white px-5 py-2.5 rounded-lg text-sm font-medium hover:bg-amethyst-600 transition-colors"
          >
            View Dashboard
          </Link>
        </div>
      </div>
    )
  }

  // Step 2: Deposit confirmation
  if (listingId && walletAddress) {
    return (
      <div className="p-4 md:p-6 max-w-lg mx-auto">
        <div className="flex items-center gap-3 mb-6">
          <button
            onClick={() => router.push('/seller/exchange')}
            className="p-2 rounded-lg hover:bg-slate-100 text-slate-500"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 className="text-2xl font-display font-bold text-slate-900">Deposit USDT</h1>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">{error}</div>
        )}

        <div className="bg-white rounded-xl border border-slate-200 p-6 space-y-5">
          <p className="text-sm text-slate-600">
            Send <span className="font-mono font-semibold text-slate-900">{parsedAmount} USDT</span> (TRC20) to the escrow address below:
          </p>

          {/* Wallet address display */}
          <div className="bg-slate-50 rounded-lg p-4 text-center space-y-2">
            <p className="text-xs text-slate-400 uppercase font-medium tracking-wide">Escrow Wallet Address</p>
            <p className="font-mono text-base text-slate-900 break-all leading-relaxed">{walletAddress}</p>
            <button
              onClick={copyWallet}
              className="inline-flex items-center gap-1.5 text-sm text-amethyst-600 font-medium hover:text-amethyst-700"
            >
              {copied ? (
                <>
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Copied!
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  Copy Address
                </>
              )}
            </button>
          </div>

          <p className="text-xs text-slate-400 text-center">
            For the hackathon demo, clicking &quot;Confirm Deposit&quot; simulates an instant deposit.
          </p>

          <button
            onClick={handleConfirmDeposit}
            disabled={confirming}
            className="w-full bg-amethyst-500 text-white py-3 rounded-lg font-medium hover:bg-amethyst-600 transition-colors disabled:opacity-50"
          >
            {confirming ? 'Confirming...' : 'Confirm Deposit'}
          </button>
        </div>
      </div>
    )
  }

  // Step 1: Create listing form
  return (
    <div className="p-4 md:p-6 max-w-lg mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => router.push('/seller/exchange')}
          className="p-2 rounded-lg hover:bg-slate-100 text-slate-500"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-display font-bold text-slate-900">Create USDT Listing</h1>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">{error}</div>
      )}

      <form onSubmit={handleCreate} className="space-y-6">
        <div className="bg-white rounded-xl border border-slate-200 p-4 space-y-4">
          <h2 className="text-sm font-semibold text-slate-700 uppercase tracking-wide">Listing Details</h2>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">USDT Amount</label>
            <input
              type="number"
              value={usdtAmount}
              onChange={(e) => setUsdtAmount(e.target.value)}
              min="1"
              step="0.01"
              required
              placeholder="100"
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Rate (MVR per USDT)</label>
            <input
              type="number"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              min="0.01"
              step="0.01"
              required
              placeholder="25.50"
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>

          {/* Calculated total */}
          {calculatedMVR > 0 && (
            <div className="bg-amethyst-50 rounded-lg p-3">
              <div className="flex justify-between text-sm">
                <span className="text-slate-600">Total value</span>
                <span className="font-mono font-semibold text-amethyst-700">{calculatedMVR.toFixed(2)} MVR</span>
              </div>
            </div>
          )}

          {/* Partial toggle */}
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-slate-700">Allow partial purchases</p>
              <p className="text-xs text-slate-400">Buyers can purchase less than the full amount</p>
            </div>
            <button
              type="button"
              onClick={() => setPartialAllowed(!partialAllowed)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                partialAllowed ? 'bg-amethyst-500' : 'bg-slate-300'
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  partialAllowed ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>
        </div>

        <button
          type="submit"
          disabled={creating}
          className="w-full bg-amethyst-500 text-white py-3 rounded-lg font-medium hover:bg-amethyst-600 transition-colors disabled:opacity-50"
        >
          {creating ? 'Creating...' : 'Create Listing'}
        </button>
      </form>
    </div>
  )
}

