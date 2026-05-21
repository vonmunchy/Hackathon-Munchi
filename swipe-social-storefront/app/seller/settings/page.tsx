'use client'

import { useState, useEffect } from 'react'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import { useRouter } from 'next/navigation'

export default function SettingsPage() {
  const sessionToken = useSessionToken()
  const router = useRouter()
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)

  // Form state
  const [description, setDescription] = useState('')
  const [instagramUsername, setInstagramUsername] = useState('')
  const [facebookPageUrl, setFacebookPageUrl] = useState('')
  const [whatsappNumber, setWhatsappNumber] = useState('')
  const [logoUrl, setLogoUrl] = useState('')

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const store = useQuery(api.stores.getBySlug, storeSlug ? { slug: storeSlug } : 'skip')
  const updateProfile = useMutation(api.stores.updateStoreProfile)

  // Populate form
  const [populated, setPopulated] = useState(false)
  useEffect(() => {
    if (store && !populated) {
      setDescription(store.description ?? '')
      setInstagramUsername(store.instagramUsername ?? '')
      setFacebookPageUrl(store.facebookPageUrl ?? '')
      setWhatsappNumber(store.whatsappNumber ?? '')
      setLogoUrl(store.logoUrl ?? '')
      setPopulated(true)
    }
  }, [store, populated])

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    if (!sessionToken) return

    setSaving(true)
    setSaved(false)

    try {
      await updateProfile({
        sessionToken,
        description: description || undefined,
        instagramUsername: instagramUsername || undefined,
        facebookPageUrl: facebookPageUrl || undefined,
        whatsappNumber: whatsappNumber || undefined,
        logoUrl: logoUrl || undefined,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    } catch (err) {
      console.error('Save error:', err)
    } finally {
      setSaving(false)
    }
  }

  async function handleLogout() {
    setLoggingOut(true)
    try {
      await fetch('/api/auth/logout', { method: 'POST' })
      localStorage.removeItem('storeSlug')
      localStorage.removeItem('storeName')
      router.replace('/seller/login')
    } catch {
      setLoggingOut(false)
    }
  }

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-display font-bold text-slate-900">Settings</h1>

      <form onSubmit={handleSave} className="space-y-6">
        <div className="bg-white rounded-lg shadow-sm p-4 space-y-4">
          <h2 className="text-sm font-display font-semibold text-slate-700 uppercase tracking-wide">Store Profile</h2>

          {/* Store name (readonly) */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Store Name</label>
            <input
              type="text"
              value={store?.name ?? ''}
              disabled
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm bg-slate-50 text-slate-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              placeholder="Tell customers about your store..."
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500 resize-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Logo URL</label>
            <input
              type="url"
              value={logoUrl}
              onChange={(e) => setLogoUrl(e.target.value)}
              placeholder="https://example.com/logo.png"
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4 space-y-4">
          <h2 className="text-sm font-display font-semibold text-slate-700 uppercase tracking-wide">Social Links</h2>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Instagram Username</label>
            <div className="flex">
              <span className="inline-flex items-center px-3 border border-r-0 border-slate-300 rounded-l-lg bg-slate-50 text-slate-400 text-sm">@</span>
              <input
                type="text"
                value={instagramUsername}
                onChange={(e) => setInstagramUsername(e.target.value)}
                placeholder="yourstorename"
                className="flex-1 px-3 py-2 border border-slate-300 rounded-r-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Facebook Page URL</label>
            <input
              type="url"
              value={facebookPageUrl}
              onChange={(e) => setFacebookPageUrl(e.target.value)}
              placeholder="https://facebook.com/yourpage"
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">WhatsApp Number</label>
            <input
              type="tel"
              value={whatsappNumber}
              onChange={(e) => setWhatsappNumber(e.target.value)}
              placeholder="+960 7XXXXXX"
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amethyst-500"
            />
          </div>
        </div>

        {/* Save button */}
        <button
          type="submit"
          disabled={saving}
          className="w-full bg-amethyst-500 text-white py-2.5 rounded-lg font-medium hover:bg-amethyst-600 transition-colors disabled:opacity-50"
        >
          {saving ? 'Saving...' : saved ? 'Saved!' : 'Save Changes'}
        </button>
      </form>

      {/* Logout */}
      <div className="border-t border-slate-200 pt-6">
        <button
          onClick={handleLogout}
          disabled={loggingOut}
          className="w-full bg-slate-100 text-slate-700 py-2.5 rounded-lg font-medium hover:bg-slate-200 transition-colors disabled:opacity-50"
        >
          {loggingOut ? 'Signing out...' : 'Sign Out'}
        </button>
      </div>
    </div>
  )
}
