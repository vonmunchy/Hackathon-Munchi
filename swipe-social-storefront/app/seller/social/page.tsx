'use client'

import { useState, useEffect } from 'react'
import { useQuery } from 'convex/react'
import { api } from '@/convex/_generated/api'
import {
  generateInstagramCaption,
  generateFacebookCaption,
  generateHashtags,
} from '@/lib/caption-templates'
import type { Id } from '@/convex/_generated/dataModel'

interface MetaStatus {
  facebook: { connected: boolean; pageId?: string; pageName?: string }
  instagram: { connected: boolean; userId?: string; username?: string }
}

export default function SocialPage() {
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const [selectedProductId, setSelectedProductId] = useState<string>('')
  const [generated, setGenerated] = useState(false)
  const [instagramCaption, setInstagramCaption] = useState('')
  const [facebookCaption, setFacebookCaption] = useState('')
  const [hashtags, setHashtags] = useState('')
  const [copiedField, setCopiedField] = useState<string | null>(null)
  const [publishing, setPublishing] = useState<string | null>(null)
  const [publishResult, setPublishResult] = useState<{ platform: string; success: boolean; postUrl?: string; error?: string } | null>(null)
  const [metaStatus, setMetaStatus] = useState<MetaStatus | null>(null)
  const [statusLoading, setStatusLoading] = useState(true)

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  // Check connection status on load and after OAuth redirect
  useEffect(() => {
    async function checkStatus() {
      try {
        const res = await fetch('/api/auth/meta/status')
        if (res.ok) {
          const data = await res.json()
          setMetaStatus(data)
        }
      } catch {
        // Ignore
      } finally {
        setStatusLoading(false)
      }
    }
    checkStatus()

    // Check URL params for OAuth results
    const params = new URLSearchParams(window.location.search)
    if (params.get('fb_connected')) {
      const pageName = params.get('fb_page') || 'Facebook Page'
      setPublishResult({ platform: 'facebook', success: true, error: undefined, postUrl: undefined })
      setPublishResult({ platform: 'connect', success: true, error: `Connected to ${pageName}!` })
    }
    if (params.get('fb_error')) {
      setPublishResult({ platform: 'connect', success: false, error: `Facebook: ${params.get('fb_error')}` })
    }
    if (params.get('ig_connected')) {
      const igUser = params.get('ig_user') || 'Instagram'
      setPublishResult({ platform: 'connect', success: true, error: `Connected to @${igUser}!` })
    }
    if (params.get('ig_error')) {
      setPublishResult({ platform: 'connect', success: false, error: `Instagram: ${params.get('ig_error')}` })
    }
    // Clean up URL params
    if (params.toString()) {
      window.history.replaceState({}, '', '/seller/social')
    }
  }, [])

  const data = useQuery(
    api.products.listByStoreSlugWithVariants,
    storeSlug ? { storeSlug } : 'skip',
  )

  const activeProducts = (data?.products ?? []).filter((p) => p.status === 'active')
  const store = data?.store

  const selectedProduct = useQuery(
    api.products.getWithVariants,
    selectedProductId ? { productId: selectedProductId as Id<'products'> } : 'skip',
  )

  function handleGenerate() {
    if (!selectedProduct || !store) return

    const storefrontUrl = `${window.location.origin}/shop/${store.slug}`

    setInstagramCaption(
      generateInstagramCaption(selectedProduct, store, selectedProduct.variants, storefrontUrl),
    )
    setFacebookCaption(
      generateFacebookCaption(selectedProduct, store, selectedProduct.variants, storefrontUrl),
    )
    setHashtags(generateHashtags(selectedProduct, store))
    setGenerated(true)
    setPublishResult(null)
  }

  async function handleCopy(text: string, field: string) {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedField(field)
      setTimeout(() => setCopiedField(null), 2000)
    } catch {
      // Fallback
    }
  }

  async function handlePublish(platform: 'facebook' | 'instagram') {
    setPublishing(platform)
    setPublishResult(null)
    try {
      const fullCaption = (platform === 'facebook' ? facebookCaption : instagramCaption) + '\n\n' + hashtags
      const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
      const storefrontUrl = `${window.location.origin}/shop/${store?.slug}`

      // Get the product's first image URL (Convex Storage public URLs work from anywhere)
      const productImage = selectedProduct?.imageUrls?.[0]
      const imageUrl = productImage?.startsWith('http') ? productImage : undefined

      const res = await fetch('/api/meta/publish', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          platform,
          message: fullCaption,
          ...(!isLocalhost && { link: `${storefrontUrl}/product/${selectedProductId}` }),
          ...(imageUrl && { imageUrl }),
        }),
      })
      const data = await res.json()
      if (res.ok) {
        setPublishResult({ platform, success: true, postUrl: data.postUrl })
      } else {
        setPublishResult({ platform, success: false, error: data.error })
      }
    } catch {
      setPublishResult({ platform, success: false, error: 'Network error' })
    } finally {
      setPublishing(null)
    }
  }

  const fbConnected = metaStatus?.facebook?.connected ?? false
  const igConnected = metaStatus?.instagram?.connected ?? false

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-display font-bold text-slate-900">Social Captions</h1>
        <p className="text-slate-500 text-sm mt-1">Generate captions and publish to your social media</p>
      </div>

      {/* Connected Accounts Section */}
      <div className="bg-white rounded-lg shadow-sm p-4 space-y-3">
        <h2 className="text-sm font-medium text-slate-700">Connected Accounts</h2>

        {/* Connection status / error message */}
        {publishResult?.platform === 'connect' && (
          <div className={`text-sm rounded-lg p-3 ${publishResult.success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-600'}`}>
            {publishResult.error}
          </div>
        )}

        {statusLoading ? (
          <div className="space-y-2">
            <div className="bg-gray-100 rounded-lg p-3 animate-pulse h-12" />
          </div>
        ) : fbConnected ? (
          <div className="space-y-2">
            {/* Facebook Page */}
            <div className="flex items-center gap-2 bg-blue-50 border border-blue-200 rounded-lg px-4 py-3">
              <svg className="w-5 h-5 text-[#1877F2]" fill="currentColor" viewBox="0 0 24 24">
                <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
              </svg>
              <div className="flex-1">
                <span className="text-sm font-medium text-blue-800">{metaStatus?.facebook?.pageName || 'Facebook Page'}</span>
                <span className="text-xs text-blue-600 ml-2">Connected</span>
              </div>
              <svg className="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>

            {/* Instagram (linked via Meta Business) */}
            {igConnected ? (
              <div className="flex items-center gap-2 bg-purple-50 border border-purple-200 rounded-lg px-4 py-3">
                <svg className="w-5 h-5 text-[#E4405F]" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zM12 0C8.741 0 8.333.014 7.053.072 2.695.272.273 2.69.073 7.052.014 8.333 0 8.741 0 12c0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98C8.333 23.986 8.741 24 12 24c3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 100 12.324 6.162 6.162 0 000-12.324zM12 16a4 4 0 110-8 4 4 0 010 8zm6.406-11.845a1.44 1.44 0 100 2.881 1.44 1.44 0 000-2.881z" />
                </svg>
                <div className="flex-1">
                  <span className="text-sm font-medium text-purple-800">@{metaStatus?.instagram?.username || 'Instagram'}</span>
                  <span className="text-xs text-purple-600 ml-2">Connected</span>
                </div>
                <svg className="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            ) : (
              <p className="text-xs text-slate-400 px-1">
                Instagram not linked. Link an Instagram Business account to your Facebook Page in Meta Business Suite to publish to Instagram.
              </p>
            )}
          </div>
        ) : (
          <a
            href="/api/auth/meta/facebook"
            className="w-full flex items-center justify-center gap-2 bg-[#0082FB] text-white rounded-lg px-4 py-3 font-medium text-sm hover:bg-[#006FDB] transition-colors"
          >
            Continue with Meta Business
          </a>
        )}
      </div>

      {/* Product Selector */}
      <div className="bg-white rounded-lg shadow-sm p-4 space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Select Product</label>
          <select
            value={selectedProductId}
            onChange={(e) => {
              setSelectedProductId(e.target.value)
              setGenerated(false)
            }}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
          >
            <option value="">Choose a product...</option>
            {activeProducts.map((p) => (
              <option key={p._id} value={p._id}>{p.name}</option>
            ))}
          </select>
        </div>

        <button
          onClick={handleGenerate}
          disabled={!selectedProductId || !selectedProduct}
          className="w-full bg-amethyst-500 text-white py-2.5 rounded-lg font-medium hover:bg-amethyst-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Generate Captions
        </button>
      </div>

      {/* Generated Captions */}
      {generated && (
        <div className="space-y-4">
          {/* Instagram */}
          <CaptionCard
            title="Instagram"
            icon={
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zM12 0C8.741 0 8.333.014 7.053.072 2.695.272.273 2.69.073 7.052.014 8.333 0 8.741 0 12c0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98C8.333 23.986 8.741 24 12 24c3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 100 12.324 6.162 6.162 0 000-12.324zM12 16a4 4 0 110-8 4 4 0 010 8zm6.406-11.845a1.44 1.44 0 100 2.881 1.44 1.44 0 000-2.881z" />
              </svg>
            }
            content={instagramCaption}
            copied={copiedField === 'instagram'}
            onCopy={() => handleCopy(instagramCaption, 'instagram')}
          />

          {/* Facebook */}
          <CaptionCard
            title="Facebook"
            icon={
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
              </svg>
            }
            content={facebookCaption}
            copied={copiedField === 'facebook'}
            onCopy={() => handleCopy(facebookCaption, 'facebook')}
          />

          {/* Publish to Facebook */}
          <div className="bg-white rounded-lg shadow-sm p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2 text-slate-700">
                <svg className="w-5 h-5 text-[#1877F2]" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
                </svg>
                <span className="font-medium text-sm">Publish to Facebook Page</span>
              </div>
            </div>
            {fbConnected ? (
              <>
                <p className="text-xs text-slate-500 mb-3">
                  Post to <strong>{metaStatus?.facebook?.pageName}</strong>
                </p>
                <button
                  onClick={() => handlePublish('facebook')}
                  disabled={publishing === 'facebook'}
                  className="w-full bg-[#1877F2] text-white py-2.5 rounded-lg font-medium hover:bg-[#166FE5] transition-colors disabled:opacity-50"
                >
                  {publishing === 'facebook' ? 'Publishing...' : 'Publish to Facebook'}
                </button>
              </>
            ) : (
              <p className="text-xs text-slate-500">
                <a href="/api/auth/meta/facebook" className="text-[#1877F2] underline font-medium">Login with Facebook</a> to publish directly to your Page
              </p>
            )}
            {publishResult?.platform === 'facebook' && (
              <div className={`mt-3 text-sm rounded-lg p-3 ${publishResult.success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-600'}`}>
                {publishResult.success ? (
                  <>Published! <a href={publishResult.postUrl} target="_blank" rel="noopener noreferrer" className="underline">View post</a></>
                ) : (
                  publishResult.error
                )}
              </div>
            )}
          </div>

          {/* Publish to Instagram */}
          <div className="bg-white rounded-lg shadow-sm p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2 text-slate-700">
                <svg className="w-5 h-5 text-[#E4405F]" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zM12 0C8.741 0 8.333.014 7.053.072 2.695.272.273 2.69.073 7.052.014 8.333 0 8.741 0 12c0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98C8.333 23.986 8.741 24 12 24c3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 100 12.324 6.162 6.162 0 000-12.324zM12 16a4 4 0 110-8 4 4 0 010 8zm6.406-11.845a1.44 1.44 0 100 2.881 1.44 1.44 0 000-2.881z" />
                </svg>
                <span className="font-medium text-sm">Publish to Instagram</span>
              </div>
            </div>
            {igConnected ? (
              <>
                <p className="text-xs text-slate-500 mb-3">
                  Post to <strong>@{metaStatus?.instagram?.username}</strong>
                </p>
                <button
                  onClick={() => handlePublish('instagram')}
                  disabled={publishing === 'instagram'}
                  className="w-full bg-gradient-to-r from-[#833AB4] via-[#FD1D1D] to-[#F77737] text-white py-2.5 rounded-lg font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
                >
                  {publishing === 'instagram' ? 'Publishing...' : 'Publish to Instagram'}
                </button>
              </>
            ) : (
              <p className="text-xs text-slate-500">
                <a href="/api/auth/meta/instagram" className="text-[#E4405F] underline font-medium">Login with Instagram</a> to publish directly to your account
              </p>
            )}
            {publishResult?.platform === 'instagram' && (
              <div className={`mt-3 text-sm rounded-lg p-3 ${publishResult.success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-600'}`}>
                {publishResult.success ? 'Published to Instagram!' : publishResult.error}
              </div>
            )}
          </div>

          {/* Hashtags */}
          <CaptionCard
            title="Hashtags"
            icon={
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
              </svg>
            }
            content={hashtags}
            copied={copiedField === 'hashtags'}
            onCopy={() => handleCopy(hashtags, 'hashtags')}
          />
        </div>
      )}
    </div>
  )
}

function CaptionCard({
  title,
  icon,
  content,
  copied,
  onCopy,
}: {
  title: string
  icon: React.ReactNode
  content: string
  copied: boolean
  onCopy: () => void
}) {
  return (
    <div className="bg-white rounded-lg shadow-sm p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2 text-slate-700">
          {icon}
          <span className="font-medium text-sm">{title}</span>
        </div>
        <button
          onClick={onCopy}
          className={`text-xs font-medium px-3 py-1 rounded-full transition-colors ${
            copied
              ? 'bg-green-100 text-green-700'
              : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
          }`}
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <pre className="text-sm text-slate-800 whitespace-pre-wrap font-body leading-relaxed">
        {content}
      </pre>
    </div>
  )
}
