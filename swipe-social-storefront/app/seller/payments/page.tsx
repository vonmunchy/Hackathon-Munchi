'use client'

import { useState, useEffect } from 'react'
import { useQuery } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { formatMVR } from '@/lib/format'
import type { Id } from '@/convex/_generated/dataModel'

export default function CreatePaymentLinkPage() {
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const [selectedProductId, setSelectedProductId] = useState<string>('')
  const [selectedVariantId, setSelectedVariantId] = useState<string>('')
  const [quantity, setQuantity] = useState(1)
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [deliveryLocation, setDeliveryLocation] = useState("Male'")
  const [deliveryAddress, setDeliveryAddress] = useState('')
  const [deliveryTime, setDeliveryTime] = useState('')
  const [loading, setLoading] = useState(false)
  const [paymentLink, setPaymentLink] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [mode, setMode] = useState<'full' | 'quick'>('quick')

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const data = useQuery(
    api.products.listByStoreSlugWithVariants,
    storeSlug ? { storeSlug } : 'skip',
  )

  const activeProducts = (data?.products ?? []).filter((p) => p.status === 'active')

  const selectedProduct = activeProducts.find((p) => p._id === selectedProductId)
  const variants = selectedProduct?.variants ?? []
  const selectedVariant = variants.find((v) => v._id === selectedVariantId)
  const unitPrice = selectedVariant?.priceOverride ?? selectedProduct?.basePrice ?? 0
  const totalAmount = unitPrice * quantity

  async function handleCreate() {
    if (!selectedProductId || !selectedVariantId || !customerName || !customerPhone || !deliveryAddress) {
      setError('Please fill in all required fields')
      return
    }

    setLoading(true)
    setError(null)
    setPaymentLink(null)

    try {
      const res = await fetch('/api/checkout/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          productId: selectedProductId,
          variantId: selectedVariantId,
          quantity,
          customerName,
          customerPhone,
          deliveryLocation,
          deliveryAddress,
          deliveryTimePreference: deliveryTime || undefined,
        }),
      })

      const result = await res.json()

      if (!res.ok) {
        setError(result.error || 'Failed to create payment link')
        return
      }

      const baseUrl = window.location.origin
      const link = `${baseUrl}/checkout/${result.orderId}?token=${result.accessToken}`
      setPaymentLink(link)
    } catch {
      setError('Network error. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    if (!paymentLink) return
    try {
      await navigator.clipboard.writeText(paymentLink)
      setCopied(true)
      setTimeout(() => setCopied(false), 2500)
    } catch {
      // fallback
    }
  }

  function handleQuickLink() {
    if (!selectedProductId || !selectedVariantId || !storeSlug) return
    const baseUrl = window.location.origin
    const link = `${baseUrl}/shop/${storeSlug}/product/${selectedProductId}?variant=${selectedVariantId}&qty=${quantity}`
    setPaymentLink(link)
  }

  function handleReset() {
    setPaymentLink(null)
    setCustomerName('')
    setCustomerPhone('')
    setDeliveryAddress('')
    setDeliveryTime('')
    setCopied(false)
    setError(null)
  }

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-display font-bold text-slate-900">Create Payment Link</h1>
        <p className="text-slate-500 text-sm mt-1">
          Create a Swipe payment link to send to your buyer in Instagram or Facebook chat
        </p>
      </div>

      {/* Mode toggle */}
      {!paymentLink && (
        <div className="flex rounded-lg border border-slate-200 bg-slate-50 p-1">
          <button
            onClick={() => setMode('quick')}
            className={`flex-1 rounded-md py-2 text-sm font-medium transition-colors ${mode === 'quick' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}
          >
            Quick Link
          </button>
          <button
            onClick={() => setMode('full')}
            className={`flex-1 rounded-md py-2 text-sm font-medium transition-colors ${mode === 'full' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}
          >
            Full Order
          </button>
        </div>
      )}

      {paymentLink ? (
        /* Success — show the link to copy */
        <div className="space-y-4">
          <div className="rounded-xl border-2 border-amethyst-200 bg-amethyst-50 p-5 space-y-3">
            <div className="flex items-center gap-2 text-amethyst-700">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
              </svg>
              <span className="font-display font-semibold">Payment link created!</span>
            </div>

            <p className="text-sm text-slate-600">
              Send this link to <span className="font-medium text-slate-800">{customerName}</span> in your chat. They'll pay <span className="font-mono font-medium text-amethyst-700">{formatMVR(totalAmount)}</span> for {selectedProduct?.name}.
            </p>

            <div className="rounded-lg bg-white border border-slate-200 p-3">
              <p className="text-xs text-slate-500 mb-1">Payment Link</p>
              <p className="text-sm font-mono text-slate-700 break-all">{paymentLink}</p>
            </div>

            <button
              onClick={handleCopy}
              className={`w-full rounded-xl py-3 text-sm font-semibold transition-colors ${
                copied
                  ? 'bg-amethyst-100 text-amethyst-700'
                  : 'bg-amethyst-500 text-white hover:bg-amethyst-600'
              }`}
            >
              {copied ? 'Copied! Paste it in the chat' : 'Copy Payment Link'}
            </button>
          </div>

          {/* Message template to send along with the link */}
          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <p className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-2">
              Suggested message to send with the link
            </p>
            <p className="text-sm text-slate-700 whitespace-pre-wrap">
              {`Hi ${customerName}! Here's your payment link for ${selectedProduct?.name}${selectedVariant ? ` (${selectedVariant.variantName})` : ''}:\n\n${paymentLink}\n\nTotal: ${formatMVR(totalAmount)}\n\nPay securely via Swipe. You'll get a confirmation once payment is complete!`}
            </p>
            <button
              onClick={async () => {
                const msg = `Hi ${customerName}! Here's your payment link for ${selectedProduct?.name}${selectedVariant ? ` (${selectedVariant.variantName})` : ''}:\n\n${paymentLink}\n\nTotal: ${formatMVR(totalAmount)}\n\nPay securely via Swipe. You'll get a confirmation once payment is complete!`
                await navigator.clipboard.writeText(msg)
              }}
              className="mt-3 w-full rounded-lg border border-slate-300 bg-white py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors"
            >
              Copy Full Message
            </button>
          </div>

          {/* Share buttons */}
          <div className="flex gap-3">
            <a
              href={`https://wa.me/?text=${encodeURIComponent(`Hi ${customerName}! Here's your payment link for ${selectedProduct?.name}:\n\n${paymentLink}\n\nTotal: ${formatMVR(totalAmount)}\n\nPay securely via Swipe!`)}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-[#25D366] py-3 text-sm font-semibold text-white hover:opacity-90 transition-opacity"
            >
              <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" />
              </svg>
              Share via WhatsApp
            </a>
          </div>

          <button
            onClick={handleReset}
            className="w-full rounded-xl border border-slate-300 bg-white py-3 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors"
          >
            Create Another Link
          </button>
        </div>
      ) : (
        /* Form to create the payment link */
        <div className="space-y-4">
          {error && (
            <div className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
              {error}
            </div>
          )}

          {/* Product selection */}
          <div className="bg-white rounded-xl shadow-sm p-4 space-y-4">
            <h2 className="text-sm font-display font-semibold text-slate-700 uppercase tracking-wide">Product</h2>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Select Product</label>
              <select
                value={selectedProductId}
                onChange={(e) => {
                  setSelectedProductId(e.target.value)
                  setSelectedVariantId('')
                  setQuantity(1)
                }}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              >
                <option value="">Choose a product...</option>
                {activeProducts.map((p) => (
                  <option key={p._id} value={p._id}>{p.name} — {formatMVR(p.basePrice)}</option>
                ))}
              </select>
            </div>

            {variants.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Variant</label>
                <select
                  value={selectedVariantId}
                  onChange={(e) => setSelectedVariantId(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
                >
                  <option value="">Choose variant...</option>
                  {variants.map((v) => (
                    <option key={v._id} value={v._id}>
                      {v.variantName} — {v.stockAvailable} in stock
                      {v.priceOverride ? ` — ${formatMVR(v.priceOverride)}` : ''}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {selectedVariantId && (
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Quantity</label>
                <input
                  type="number"
                  min={1}
                  max={selectedVariant?.stockAvailable ?? 99}
                  value={quantity}
                  onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                  className="w-20 px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
                />
              </div>
            )}

            {selectedVariantId && (
              <div className="pt-2 border-t border-slate-200">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">Total</span>
                  <span className="font-mono font-semibold text-amethyst-700">{formatMVR(totalAmount)}</span>
                </div>
              </div>
            )}
          </div>

          {/* Buyer details — only in full mode */}
          {mode === 'full' && <div className="bg-white rounded-xl shadow-sm p-4 space-y-4">
            <h2 className="text-sm font-display font-semibold text-slate-700 uppercase tracking-wide">Buyer Details</h2>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Buyer Name</label>
              <input
                type="text"
                value={customerName}
                onChange={(e) => setCustomerName(e.target.value)}
                placeholder="Name from the chat"
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Phone Number</label>
              <input
                type="tel"
                value={customerPhone}
                onChange={(e) => setCustomerPhone(e.target.value)}
                placeholder="7XXXXXX or 9XXXXXX"
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Delivery Location</label>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 text-sm text-slate-700">
                  <input
                    type="radio"
                    name="location"
                    checked={deliveryLocation === "Male'"}
                    onChange={() => setDeliveryLocation("Male'")}
                    className="accent-amethyst-500"
                  />
                  Male&apos;
                </label>
                <label className="flex items-center gap-2 text-sm text-slate-700">
                  <input
                    type="radio"
                    name="location"
                    checked={deliveryLocation === "Hulhumale'"}
                    onChange={() => setDeliveryLocation("Hulhumale'")}
                    className="accent-amethyst-500"
                  />
                  Hulhumale&apos;
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Delivery Address</label>
              <textarea
                value={deliveryAddress}
                onChange={(e) => setDeliveryAddress(e.target.value)}
                placeholder="Building, floor, road..."
                rows={2}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500 resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                Delivery Time <span className="text-slate-400 font-normal">(optional)</span>
              </label>
              <input
                type="text"
                value={deliveryTime}
                onChange={(e) => setDeliveryTime(e.target.value)}
                placeholder="e.g. After 5pm, Weekend morning"
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
            </div>
          </div>}

          {mode === 'quick' ? (
            <button
              onClick={handleQuickLink}
              disabled={!selectedProductId || !selectedVariantId}
              className="w-full rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              {`Generate Product Link — ${formatMVR(totalAmount)}`}
            </button>
          ) : (
            <button
              onClick={handleCreate}
              disabled={loading || !selectedProductId || !selectedVariantId || !customerName || !customerPhone || !deliveryAddress}
              className="w-full rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              {loading ? 'Creating...' : `Create Payment Link — ${formatMVR(totalAmount)}`}
            </button>
          )}
        </div>
      )}
    </div>
  )
}
