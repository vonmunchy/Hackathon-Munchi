'use client'

import { useState, useEffect } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import type { Id } from '@/convex/_generated/dataModel'

interface VariantForm {
  id?: string
  variantName: string
  priceOverride: string
  stockAvailable: string
}

const CATEGORIES = ['Fashion', 'Gifts', 'Accessories', 'Food/Gifts', 'Other']

export default function NewProductPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const editId = searchParams.get('edit')

  const sessionToken = useSessionToken()
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Form state
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [basePrice, setBasePrice] = useState('')
  const [category, setCategory] = useState('Other')
  const [status, setStatus] = useState('draft')
  const [imageUrls, setImageUrls] = useState<string[]>([''])
  const [variants, setVariants] = useState<VariantForm[]>([
    { variantName: '', priceOverride: '', stockAvailable: '0' },
  ])

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const store = useQuery(api.stores.getBySlug, storeSlug ? { slug: storeSlug } : 'skip')

  // Load existing product data if editing
  const existingProduct = useQuery(
    api.products.getWithVariants,
    editId ? { productId: editId as Id<'products'> } : 'skip',
  )

  const createProduct = useMutation(api.products.create)
  const updateProduct = useMutation(api.products.update)
  const addVariant = useMutation(api.products.addVariant)
  const removeVariantMut = useMutation(api.products.removeVariant)
  const updateVariantMut = useMutation(api.products.updateVariant)

  // Populate form when editing
  const [populated, setPopulated] = useState(false)
  useEffect(() => {
    if (existingProduct && !populated) {
      setName(existingProduct.name)
      setDescription(existingProduct.description ?? '')
      setBasePrice(String(existingProduct.basePrice))
      setCategory(existingProduct.category ?? 'Other')
      setStatus(existingProduct.status)
      setImageUrls(
        existingProduct.imageUrls.length > 0
          ? [...existingProduct.imageUrls]
          : [''],
      )
      setVariants(
        existingProduct.variants.map((v) => ({
          id: v._id,
          variantName: v.variantName,
          priceOverride: v.priceOverride != null ? String(v.priceOverride) : '',
          stockAvailable: String(v.stockAvailable),
        })),
      )
      setPopulated(true)
    }
  }, [existingProduct, populated])

  function addImageField() {
    if (imageUrls.length < 5) {
      setImageUrls([...imageUrls, ''])
    }
  }

  function updateImageUrl(index: number, value: string) {
    const updated = [...imageUrls]
    updated[index] = value
    setImageUrls(updated)
  }

  function removeImageField(index: number) {
    setImageUrls(imageUrls.filter((_, i) => i !== index))
  }

  function addVariantField() {
    setVariants([...variants, { variantName: '', priceOverride: '', stockAvailable: '0' }])
  }

  function updateVariantField(index: number, field: keyof VariantForm, value: string) {
    const updated = [...variants]
    updated[index] = { ...updated[index], [field]: value }
    setVariants(updated)
  }

  function removeVariantField(index: number) {
    setVariants(variants.filter((_, i) => i !== index))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!store || !sessionToken) return

    setSaving(true)
    setError(null)

    try {
      const filteredUrls = imageUrls.filter((u) => u.trim() !== '')
      const parsedPrice = parseFloat(basePrice)

      if (isNaN(parsedPrice) || parsedPrice <= 0) {
        setError('Please enter a valid base price')
        setSaving(false)
        return
      }

      if (editId) {
        // Update product
        await updateProduct({
          sessionToken,
          productId: editId as Id<'products'>,
          name,
          description: description || undefined,
          basePrice: parsedPrice,
          currency: 'MVR',
          category,
          imageUrls: filteredUrls,
          status,
        })

        // Handle variants: for editing, sync existing variants
        if (existingProduct) {
          const existingIds = new Set<string>(existingProduct.variants.map((v) => v._id as string))
          const formIds = new Set<string>(variants.filter((v) => v.id).map((v) => v.id!))

          // Remove deleted variants
          for (const ev of existingProduct.variants) {
            if (!formIds.has(ev._id)) {
              await removeVariantMut({ sessionToken, variantId: ev._id as Id<'productVariants'> })
            }
          }

          // Update existing and add new
          for (const v of variants) {
            if (v.id && existingIds.has(v.id)) {
              await updateVariantMut({
                sessionToken,
                variantId: v.id as Id<'productVariants'>,
                variantName: v.variantName,
                priceOverride: v.priceOverride ? parseFloat(v.priceOverride) : undefined,
                stockAvailable: parseInt(v.stockAvailable, 10) || 0,
              })
            } else if (!v.id && v.variantName.trim()) {
              await addVariant({
                sessionToken,
                productId: editId as Id<'products'>,
                variantName: v.variantName,
                priceOverride: v.priceOverride ? parseFloat(v.priceOverride) : undefined,
                stockAvailable: parseInt(v.stockAvailable, 10) || 0,
              })
            }
          }
        }
      } else {
        // Create new product with variants
        const validVariants = variants
          .filter((v) => v.variantName.trim())
          .map((v) => ({
            variantName: v.variantName,
            priceOverride: v.priceOverride ? parseFloat(v.priceOverride) : undefined,
            stockAvailable: parseInt(v.stockAvailable, 10) || 0,
            stockSold: 0,
          }))

        await createProduct({
          sessionToken,
          name,
          description: description || undefined,
          basePrice: parsedPrice,
          currency: 'MVR',
          category,
          imageUrls: filteredUrls,
          status,
          createdAt: Date.now(),
          variants: validVariants,
        })
      }

      router.push('/seller/products')
    } catch (err) {
      console.error('Save product error:', err)
      setError('Failed to save product. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  const isEditing = !!editId
  const pageTitle = isEditing ? 'Edit Product' : 'New Product'

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => router.push('/seller/products')}
          className="p-2 rounded-lg hover:bg-slate-100 text-slate-500"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-display font-bold text-slate-900">{pageTitle}</h1>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <div className="bg-white rounded-lg shadow-sm p-4 space-y-4">
          <h2 className="text-sm font-semibold text-slate-700 uppercase tracking-wide">Basic Info</h2>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500 resize-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Base Price (MVR)</label>
              <input
                type="number"
                value={basePrice}
                onChange={(e) => setBasePrice(e.target.value)}
                min="0"
                step="0.01"
                required
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Category</label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              >
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            >
              <option value="draft">Draft</option>
              <option value="active">Active</option>
              <option value="archived">Archived</option>
            </select>
          </div>
        </div>

        {/* Images */}
        <div className="bg-white rounded-lg shadow-sm p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-700 uppercase tracking-wide">Images</h2>
            {imageUrls.length < 5 && (
              <button
                type="button"
                onClick={addImageField}
                className="text-xs text-amethyst-500 font-medium hover:text-amethyst-600"
              >
                + Add Image
              </button>
            )}
          </div>

          {imageUrls.map((url, i) => (
            <div key={i} className="flex gap-2">
              <input
                type="url"
                value={url}
                onChange={(e) => updateImageUrl(i, e.target.value)}
                placeholder="https://example.com/image.jpg"
                className="flex-1 px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
              {imageUrls.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeImageField(i)}
                  className="p-2 text-slate-400 hover:text-red-500"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              )}
            </div>
          ))}
        </div>

        {/* Variants */}
        <div className="bg-white rounded-lg shadow-sm p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-700 uppercase tracking-wide">Variants</h2>
            <button
              type="button"
              onClick={addVariantField}
              className="text-xs text-amethyst-500 font-medium hover:text-amethyst-600"
            >
              + Add Variant
            </button>
          </div>

          {variants.map((v, i) => (
            <div key={i} className="border border-slate-200 rounded-lg p-3 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-slate-400">Variant {i + 1}</span>
                {variants.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeVariantField(i)}
                    className="text-xs text-red-400 hover:text-red-600"
                  >
                    Remove
                  </button>
                )}
              </div>
              <input
                type="text"
                value={v.variantName}
                onChange={(e) => updateVariantField(i, 'variantName', e.target.value)}
                placeholder="Variant name (e.g. Red / Large)"
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-xs text-slate-400 mb-0.5">Price Override (MVR)</label>
                  <input
                    type="number"
                    value={v.priceOverride}
                    onChange={(e) => updateVariantField(i, 'priceOverride', e.target.value)}
                    placeholder="Optional"
                    min="0"
                    step="0.01"
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
                  />
                </div>
                <div>
                  <label className="block text-xs text-slate-400 mb-0.5">Stock</label>
                  <input
                    type="number"
                    value={v.stockAvailable}
                    onChange={(e) => updateVariantField(i, 'stockAvailable', e.target.value)}
                    min="0"
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
                  />
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Submit */}
        <button
          type="submit"
          disabled={saving}
          className="w-full bg-amethyst-500 text-white py-3 rounded-lg font-medium hover:bg-amethyst-600 transition-colors disabled:opacity-50"
        >
          {saving ? 'Saving...' : isEditing ? 'Update Product' : 'Save Product'}
        </button>
      </form>
    </div>
  )
}
