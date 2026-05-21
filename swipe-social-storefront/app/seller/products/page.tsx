'use client'

import { useState, useEffect } from 'react'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import { formatMVR } from '@/lib/format'
import { useIsMobile } from '@/lib/use-device'
import Link from 'next/link'

const statusBadgeColors: Record<string, string> = {
  active: 'bg-amethyst-100 text-amethyst-700',
  draft: 'bg-slate-200 text-slate-600',
  archived: 'bg-red-50 text-red-600',
}

export default function ProductsPage() {
  const sessionToken = useSessionToken()
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const isMobile = useIsMobile()

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const data = useQuery(
    api.products.listByStoreSlugWithVariants,
    storeSlug ? { storeSlug } : 'skip',
  )

  const products = data?.products ?? []
  const importTestProducts = useMutation(api.products.importTestProducts)
  const [importing, setImporting] = useState(false)

  async function handleImportTest() {
    if (!sessionToken) return
    setImporting(true)
    try {
      await importTestProducts({ sessionToken })
    } catch { /* ignore */ }
    finally { setImporting(false) }
  }

  return (
    <div className="p-4 md:p-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-display font-bold text-slate-900">Products</h1>
        <Link
          href="/seller/products/new"
          className="inline-flex items-center gap-1.5 bg-amethyst-500 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-amethyst-600 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Product
        </Link>
      </div>

      {products.length === 0 && data !== undefined ? (
        <div className="text-center py-12 text-slate-400">
          <svg className="w-12 h-12 mx-auto mb-3 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
          <p className="text-lg font-medium">No products yet</p>
          <p className="text-sm mt-1">Add your first product or import test products to get started</p>
          <button
            onClick={handleImportTest}
            disabled={importing}
            className="mt-4 inline-flex items-center gap-1.5 bg-ruby-500 text-white px-5 py-2.5 rounded-lg text-sm font-medium hover:bg-ruby-600 transition-colors disabled:bg-slate-300"
          >
            {importing ? 'Importing...' : 'Import Test Products'}
          </button>
        </div>
      ) : data === undefined ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-100 rounded-lg h-16 animate-pulse" />
          ))}
        </div>
      ) : isMobile ? (
        /* Mobile: Card layout */
        <div className="space-y-3">
          {products.map((product) => {
            const totalStock = product.variants.reduce((s, v) => s + v.stockAvailable, 0)
            const totalSold = product.variants.reduce((s, v) => s + v.stockSold, 0)
            return (
              <Link
                key={product._id}
                href={`/seller/products/new?edit=${product._id}`}
                className="block bg-white rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start gap-3">
                  {product.imageUrls[0] ? (
                    <img
                      src={product.imageUrls[0]}
                      alt={product.name}
                      className="w-14 h-14 rounded-lg object-cover bg-slate-100"
                    />
                  ) : (
                    <div className="w-14 h-14 rounded-lg bg-slate-100 flex items-center justify-center text-slate-400 text-xs">
                      No img
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-start justify-between">
                      <p className="font-medium text-slate-900 truncate">{product.name}</p>
                      <span className={`ml-2 shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ${statusBadgeColors[product.status] ?? 'bg-slate-100 text-slate-500'}`}>
                        {product.status}
                      </span>
                    </div>
                    <p className="text-sm font-mono text-amethyst-700 mt-0.5">{formatMVR(product.basePrice)}</p>
                    <div className="flex gap-3 text-xs text-slate-400 mt-1">
                      <span>{product.variants.length} variant{product.variants.length !== 1 ? 's' : ''}</span>
                      <span>Stock: {totalStock}</span>
                      <span>Sold: {totalSold}</span>
                    </div>
                  </div>
                </div>
              </Link>
            )
          })}
        </div>
      ) : (
        /* Desktop: Table layout */
        <div className="bg-white rounded-lg shadow-sm overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-200 text-left text-slate-500 bg-slate-50">
                <th className="px-4 py-3 font-medium">Product</th>
                <th className="px-4 py-3 font-medium">Price</th>
                <th className="px-4 py-3 font-medium">Variants</th>
                <th className="px-4 py-3 font-medium">Stock</th>
                <th className="px-4 py-3 font-medium">Sold</th>
                <th className="px-4 py-3 font-medium">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {products.map((product) => {
                const totalStock = product.variants.reduce((s, v) => s + v.stockAvailable, 0)
                const totalSold = product.variants.reduce((s, v) => s + v.stockSold, 0)
                return (
                  <tr key={product._id} className="hover:bg-slate-50">
                    <td className="px-4 py-3">
                      <Link
                        href={`/seller/products/new?edit=${product._id}`}
                        className="flex items-center gap-3 hover:text-amethyst-600"
                      >
                        {product.imageUrls[0] ? (
                          <img
                            src={product.imageUrls[0]}
                            alt={product.name}
                            className="w-10 h-10 rounded-lg object-cover bg-slate-100"
                          />
                        ) : (
                          <div className="w-10 h-10 rounded-lg bg-slate-100" />
                        )}
                        <span className="font-medium text-slate-900">{product.name}</span>
                      </Link>
                    </td>
                    <td className="px-4 py-3 font-mono text-amethyst-700">{formatMVR(product.basePrice)}</td>
                    <td className="px-4 py-3 text-slate-600">{product.variants.length}</td>
                    <td className="px-4 py-3 text-slate-600">{totalStock}</td>
                    <td className="px-4 py-3 text-slate-600">{totalSold}</td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusBadgeColors[product.status] ?? 'bg-slate-100 text-slate-500'}`}>
                        {product.status}
                      </span>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
