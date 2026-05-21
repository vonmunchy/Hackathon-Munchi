'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'

type SellerType = 'marketplace' | 'crypto' | 'both'

export default function OnboardingPage() {
  const sessionToken = useSessionToken()
  const router = useRouter()
  const [step, setStep] = useState(0)
  const [sellerType, setSellerType] = useState<SellerType | null>(null)
  const [storeSlug, setStoreSlug] = useState<string | null>(null)

  // Step 1 state (phone)
  const [phone, setPhone] = useState('')

  // Step 2 state (Swipe creds)
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [testResult, setTestResult] = useState<{ valid: boolean; message?: string; error?: string } | null>(null)
  const [testLoading, setTestLoading] = useState(false)

  // Step 3 state (import products)
  const [importLoading, setImportLoading] = useState(false)
  const [imported, setImported] = useState(false)

  // Crypto wallet state
  const [walletAddress, setWalletAddress] = useState('')

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const store = useQuery(api.stores.getBySlug, storeSlug ? { slug: storeSlug } : 'skip')
  const setContactPhone = useMutation(api.stores.setContactPhone)
  const setSwipeCredentials = useMutation(api.stores.setSwipeCredentials)
  const completeOnboarding = useMutation(api.stores.completeOnboarding)
  const importTestProducts = useMutation(api.products.importTestProducts)
  const setSellerTypeMutation = useMutation(api.stores.setSellerType)
  const setCryptoWallet = useMutation(api.stores.setCryptoWallet)

  // Pre-fill phone from store
  useEffect(() => {
    if (store?.whatsappNumber && !phone) {
      setPhone(store.whatsappNumber.replace(/^\+960/, ''))
    } else if (store?.contactPhone && !phone) {
      setPhone(store.contactPhone.replace(/^\+960/, ''))
    }
  }, [store, phone])

  // Pre-fill Swipe creds from store if already set
  useEffect(() => {
    if (store?.swipeClientId && !clientId) setClientId(store.swipeClientId)
    if (store?.swipeClientSecret && !clientSecret) setClientSecret(store.swipeClientSecret)
  }, [store, clientId, clientSecret])

  // Determine step flow based on sellerType
  function getStepLabels(): string[] {
    if (sellerType === 'marketplace') return ['Path', 'Phone', 'Swipe Credentials', 'Test Products']
    if (sellerType === 'crypto') return ['Path', 'Phone', 'Wallet']
    if (sellerType === 'both') return ['Path', 'Phone', 'Swipe Credentials', 'Test Products', 'Wallet']
    return ['Path']
  }

  function getTotalSteps(): number {
    return getStepLabels().length - 1 // subtract path step from displayed count
  }

  async function handleSelectPath(type: SellerType) {
    if (!sessionToken) return
    setSellerType(type)
    await setSellerTypeMutation({ sessionToken, sellerType: type })
    setStep(1)
  }

  async function handlePhoneStep() {
    if (!sessionToken || !phone) return
    await setContactPhone({ sessionToken, contactPhone: `+960${phone.replace(/^\+960/, '')}` })
    setStep(2)
  }

  async function handleTestConnection() {
    setTestLoading(true)
    setTestResult(null)
    try {
      const res = await fetch('/api/swipe/test-credentials', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clientId, clientSecret }),
      })
      const data = await res.json()
      setTestResult(data)
    } catch {
      setTestResult({ valid: false, error: 'Connection failed' })
    } finally {
      setTestLoading(false)
    }
  }

  async function handleSwipeStep() {
    if (!sessionToken || !clientId || !clientSecret) return
    await setSwipeCredentials({ sessionToken, swipeClientId: clientId, swipeClientSecret: clientSecret })
    setStep(3)
  }

  async function handleImport() {
    if (!sessionToken) return
    setImportLoading(true)
    try {
      await importTestProducts({ sessionToken })
      setImported(true)
    } catch {
      // ignore
    } finally {
      setImportLoading(false)
    }
  }

  async function handleProductsNext() {
    if (sellerType === 'both') {
      // Move to wallet step
      setStep(4)
    } else {
      await handleFinish()
    }
  }

  async function handleWalletSubmit() {
    if (!sessionToken || !isWalletValid) return
    await setCryptoWallet({ sessionToken, cryptoWalletAddress: walletAddress })
    await handleFinish()
  }

  async function handleFinish() {
    if (!sessionToken) return
    await completeOnboarding({ sessionToken })
    document.cookie = 'onboarding_complete=1; path=/'
    router.push('/seller')
  }

  // Wallet validation
  const isWalletValid = walletAddress.startsWith('T') && walletAddress.length === 34 && /^[A-Za-z0-9]+$/.test(walletAddress)
  const walletTouched = walletAddress.length > 0

  if (!storeSlug || store === undefined) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-amethyst-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  const storeName = store?.name || localStorage.getItem('storeName') || 'Your Store'

  // For crypto path, step 2 is the wallet step
  const isCryptoWalletStep = (sellerType === 'crypto' && step === 2) || (sellerType === 'both' && step === 4)
  // For crypto path, after phone we go straight to wallet
  const isSwipeStep = sellerType !== 'crypto' && step === 2
  const isProductsStep = sellerType !== 'crypto' && step === 3

  // Progress calculation (exclude step 0 from display)
  const displayStep = step
  const displayTotal = getTotalSteps()

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-10">
      <div className="w-full max-w-lg">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="font-display text-2xl font-bold text-slate-900">
            Set Up {storeName}
          </h1>
          {step > 0 && (
            <>
              <p className="mt-2 text-sm text-slate-500">Step {displayStep} of {displayTotal}</p>
              <div className="mt-4 flex gap-2">
                {Array.from({ length: displayTotal }, (_, i) => i + 1).map((s) => (
                  <div
                    key={s}
                    className={`h-1.5 flex-1 rounded-full transition-colors ${
                      s <= displayStep ? 'bg-amethyst-500' : 'bg-slate-200'
                    }`}
                  />
                ))}
              </div>
            </>
          )}
        </div>

        {/* Step 0: Path Selection */}
        {step === 0 && (
          <div className="space-y-4">
            <div className="text-center mb-6">
              <h2 className="font-display text-lg font-semibold text-slate-800">What would you like to do?</h2>
              <p className="mt-1 text-sm text-slate-500">Choose how you want to use your store</p>
            </div>

            {/* Sell Products Card */}
            <button
              onClick={() => handleSelectPath('marketplace')}
              className="w-full rounded-xl border-2 border-slate-200 bg-white p-5 text-left transition-all hover:border-amethyst-400 hover:shadow-md active:scale-[0.98]"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-amethyst-50 text-amethyst-600">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-6 w-6">
                    <path fillRule="evenodd" d="M7.5 6v.75H5.513c-.96 0-1.764.724-1.865 1.679l-1.263 12A1.875 1.875 0 004.25 22.5h15.5a1.875 1.875 0 001.865-2.071l-1.263-12a1.875 1.875 0 00-1.865-1.679H16.5V6a4.5 4.5 0 10-9 0zM12 3a3 3 0 00-3 3v.75h6V6a3 3 0 00-3-3zm-3 8.25a3 3 0 106 0v-.75a.75.75 0 011.5 0v.75a4.5 4.5 0 11-9 0v-.75a.75.75 0 011.5 0v.75z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-semibold text-slate-900">Sell Products</h3>
                  <p className="mt-0.5 text-sm text-slate-500">List and sell physical products via your storefront</p>
                </div>
              </div>
            </button>

            {/* Sell Crypto Card */}
            <button
              onClick={() => handleSelectPath('crypto')}
              className="w-full rounded-xl border-2 border-slate-200 bg-white p-5 text-left transition-all hover:border-amethyst-400 hover:shadow-md active:scale-[0.98]"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-amethyst-50 text-amethyst-600">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-6 w-6">
                    <path fillRule="evenodd" d="M12 1.5a5.25 5.25 0 00-5.25 5.25v3a3 3 0 00-3 3v6.75a3 3 0 003 3h10.5a3 3 0 003-3v-6.75a3 3 0 00-3-3v-3c0-2.9-2.35-5.25-5.25-5.25zm3.75 8.25v-3a3.75 3.75 0 10-7.5 0v3h7.5z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-semibold text-slate-900">Sell Crypto</h3>
                  <p className="mt-0.5 text-sm text-slate-500">Sell USDT securely with instant Swipe payments</p>
                </div>
              </div>
            </button>

            {/* Sell Both Card */}
            <button
              onClick={() => handleSelectPath('both')}
              className="w-full rounded-xl border-2 border-slate-200 bg-white p-5 text-left transition-all hover:border-amethyst-400 hover:shadow-md active:scale-[0.98]"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-amethyst-50 text-amethyst-600">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="h-6 w-6">
                    <path d="M11.644 1.59a.75.75 0 01.712 0l9.75 5.25a.75.75 0 010 1.32l-9.75 5.25a.75.75 0 01-.712 0l-9.75-5.25a.75.75 0 010-1.32l9.75-5.25z" />
                    <path d="M3.265 10.602l7.668 4.129a2.25 2.25 0 002.134 0l7.668-4.13 1.37.739a.75.75 0 010 1.32l-9.75 5.25a.75.75 0 01-.71 0l-9.75-5.25a.75.75 0 010-1.32l1.37-.738z" />
                    <path d="M3.265 15.602l7.668 4.129a2.25 2.25 0 002.134 0l7.668-4.13 1.37.739a.75.75 0 010 1.32l-9.75 5.25a.75.75 0 01-.71 0l-9.75-5.25a.75.75 0 010-1.32l1.37-.738z" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-display font-semibold text-slate-900">Sell Both</h3>
                  <p className="mt-0.5 text-sm text-slate-500">Access both marketplace and crypto exchange features</p>
                </div>
              </div>
            </button>
          </div>
        )}

        {/* Step 1: Contact Phone (all paths) */}
        {step === 1 && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 space-y-5">
            <div>
              <h2 className="font-display text-lg font-semibold text-slate-800">Contact Number</h2>
              <p className="mt-1 text-sm text-slate-500">
                Your phone number for customers to reach you. This will also be used for WhatsApp.
              </p>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Phone</label>
              <div className="flex">
                <span className="inline-flex items-center rounded-l-lg border border-r-0 border-slate-200 bg-slate-50 px-3 text-sm text-slate-500">
                  +960
                </span>
                <input
                  type="tel"
                  inputMode="numeric"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value.replace(/\D/g, '').slice(0, 7))}
                  placeholder="7XXXXXX"
                  className="flex-1 rounded-r-lg border border-slate-200 px-3 py-2.5 font-mono text-slate-800 placeholder:font-sans placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
                />
              </div>
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => setStep(0)}
                className="flex-1 rounded-xl border-2 border-slate-200 py-3 text-base font-semibold text-slate-600 hover:bg-slate-50"
              >
                Back
              </button>
              <button
                onClick={handlePhoneStep}
                disabled={phone.length < 7}
                className="flex-1 rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:bg-slate-300"
              >
                Next
              </button>
            </div>
          </div>
        )}

        {/* Step 2: Swipe API Key (marketplace and both paths) */}
        {isSwipeStep && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 space-y-5">
            <div>
              <h2 className="font-display text-lg font-semibold text-slate-800">Swipe API Credentials</h2>
              <p className="mt-1 text-sm text-slate-500">
                Enter your Swipe merchant credentials to receive payments.
              </p>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Client ID</label>
              <input
                type="text"
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
                placeholder="cli_..."
                className="w-full rounded-lg border border-slate-200 px-3 py-2.5 font-mono text-sm text-slate-800 placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Client Secret</label>
              <input
                type="password"
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                placeholder="sec_..."
                className="w-full rounded-lg border border-slate-200 px-3 py-2.5 font-mono text-sm text-slate-800 placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
              />
            </div>

            {/* Test Connection */}
            <button
              onClick={handleTestConnection}
              disabled={!clientId || !clientSecret || testLoading}
              className="w-full rounded-lg border-2 border-slate-200 bg-white py-2.5 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-50"
            >
              {testLoading ? 'Testing...' : 'Test Connection'}
            </button>
            {testResult && (
              <div className={`rounded-lg px-3 py-2 text-sm ${testResult.valid ? 'bg-success/10 text-success' : 'bg-error/10 text-error'}`}>
                {testResult.valid ? testResult.message : testResult.error}
              </div>
            )}

            <div className="flex gap-3">
              <button
                onClick={() => setStep(1)}
                className="flex-1 rounded-xl border-2 border-slate-200 py-3 text-base font-semibold text-slate-600 hover:bg-slate-50"
              >
                Back
              </button>
              <button
                onClick={handleSwipeStep}
                disabled={!clientId || !clientSecret}
                className="flex-1 rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:bg-slate-300"
              >
                Next
              </button>
            </div>
          </div>
        )}

        {/* Step 3: Import Test Products (marketplace and both paths) */}
        {isProductsStep && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 space-y-5">
            <div>
              <h2 className="font-display text-lg font-semibold text-slate-800">Import Test Products</h2>
              <p className="mt-1 text-sm text-slate-500">
                Get started quickly with 5 sample products. You can edit or replace them later.
              </p>
            </div>

            {/* Product preview cards */}
            <div className="space-y-2">
              {[
                { name: 'Black Abaya', price: 650, img: '/demo-products/abaya.png' },
                { name: 'Eid Gift Box', price: 450, img: '/demo-products/gift-box.png' },
                { name: 'iPhone Case', price: 120, img: '/demo-products/phone-case.png' },
                { name: 'Coral Bracelet', price: 85, img: '/demo-products/bracelet.png' },
                { name: 'Premium Dates Box', price: 280, img: '/demo-products/dates-box.png' },
              ].map((p) => (
                <div key={p.name} className="flex items-center gap-3 rounded-lg border border-slate-100 bg-slate-50 px-3 py-2">
                  <img src={p.img} alt={p.name} className="h-10 w-10 rounded-lg object-cover bg-white" />
                  <span className="flex-1 text-sm font-medium text-slate-800">{p.name}</span>
                  <span className="text-sm text-amethyst-600 font-mono">MVR {p.price}</span>
                </div>
              ))}
            </div>

            {!imported ? (
              <button
                onClick={handleImport}
                disabled={importLoading}
                className="w-full rounded-xl bg-ruby-500 py-3 text-base font-semibold text-white transition-colors hover:bg-ruby-600 disabled:bg-slate-300"
              >
                {importLoading ? 'Importing...' : 'Import All Test Products'}
              </button>
            ) : (
              <div className="rounded-lg bg-success/10 px-3 py-2.5 text-sm text-success text-center font-medium">
                5 products imported successfully!
              </div>
            )}

            <div className="flex gap-3">
              <button
                onClick={() => setStep(2)}
                className="flex-1 rounded-xl border-2 border-slate-200 py-3 text-base font-semibold text-slate-600 hover:bg-slate-50"
              >
                Back
              </button>
              <button
                onClick={handleProductsNext}
                className="flex-1 rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white transition-colors hover:bg-amethyst-600"
              >
                {sellerType === 'both' ? 'Next' : (imported ? 'Go to Dashboard' : 'Skip & Finish')}
              </button>
            </div>
          </div>
        )}

        {/* Crypto Wallet Step */}
        {isCryptoWalletStep && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 space-y-5">
            <div>
              <h2 className="font-display text-lg font-semibold text-slate-800">Your USDT Wallet Address (TRC20)</h2>
              <p className="mt-1 text-sm text-slate-500">
                This is where you will receive withdrawals. Must start with T, 34 characters.
              </p>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Wallet Address</label>
              <div className="relative">
                <input
                  type="text"
                  value={walletAddress}
                  onChange={(e) => setWalletAddress(e.target.value.replace(/[^A-Za-z0-9]/g, '').slice(0, 34))}
                  placeholder="T..."
                  className={`w-full rounded-lg border px-3 py-2.5 font-mono text-sm text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 ${
                    walletTouched
                      ? isWalletValid
                        ? 'border-green-400 focus:border-green-400 focus:ring-green-100'
                        : 'border-red-400 focus:border-red-400 focus:ring-red-100'
                      : 'border-slate-200 focus:border-amethyst-400 focus:ring-amethyst-100'
                  }`}
                />
                {walletTouched && (
                  <span className="absolute right-3 top-1/2 -translate-y-1/2">
                    {isWalletValid ? (
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="h-5 w-5 text-green-500">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
                      </svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="h-5 w-5 text-red-500">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clipRule="evenodd" />
                      </svg>
                    )}
                  </span>
                )}
              </div>
              {walletTouched && !isWalletValid && (
                <p className="mt-1.5 text-xs text-red-500">
                  {!walletAddress.startsWith('T') ? 'Address must start with T' : `Address must be exactly 34 characters (currently ${walletAddress.length})`}
                </p>
              )}
              <p className="mt-1.5 text-xs text-slate-400">
                Example: TJfKxBkqz1x5Qp7ZxkLkg1F5DwLQGpV8Br
              </p>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => setStep(sellerType === 'both' ? 3 : 1)}
                className="flex-1 rounded-xl border-2 border-slate-200 py-3 text-base font-semibold text-slate-600 hover:bg-slate-50"
              >
                Back
              </button>
              <button
                onClick={handleWalletSubmit}
                disabled={!isWalletValid}
                className="flex-1 rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:bg-slate-300"
              >
                Continue
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
