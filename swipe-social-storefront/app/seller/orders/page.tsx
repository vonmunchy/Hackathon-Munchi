'use client'

import { useState, useEffect } from 'react'
import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import { formatMVR, formatMaldivesTime, shortOrderId } from '@/lib/format'
import { OrderStatusBadge } from '@/components/seller/order-status-badge'
import { useIsMobile } from '@/lib/use-device'
import type { Id } from '@/convex/_generated/dataModel'

const TABS = ['all', 'paid', 'shipped', 'delivered', 'cancelled'] as const
type Tab = (typeof TABS)[number]

const FIVE_MINUTES = 5 * 60 * 1000

export default function OrdersPage() {
  const sessionToken = useSessionToken()
  const [storeSlug, setStoreSlug] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<Tab>('all')
  const isMobile = useIsMobile()

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
    // Mark orders as seen
    localStorage.setItem('lastSeenOrdersAt', String(Date.now()))
  }, [])

  const store = useQuery(api.stores.getBySlug, storeSlug ? { slug: storeSlug } : 'skip')
  const orders = useQuery(api.orders.listByStore, sessionToken ? { sessionToken } : 'skip')
  const fulfillOrder = useMutation(api.orders.fulfillOrder)

  // Filter out abandoned pending orders (older than 5 min)
  const now = Date.now()
  const filteredOrders = (orders ?? []).filter((order) => {
    if (order.status === 'pending' && now - order.createdAt > FIVE_MINUTES) {
      return false
    }
    if (activeTab !== 'all' && order.status !== activeTab) {
      return false
    }
    return true
  })

  async function handleFulfill(orderId: Id<'orders'>, newStatus: string) {
    if (!sessionToken) return
    try {
      await fulfillOrder({ sessionToken, orderId, newStatus })
    } catch (err) {
      console.error('Fulfill error:', err)
    }
  }

  return (
    <div className="p-4 md:p-6 max-w-5xl mx-auto">
      <h1 className="text-2xl font-display font-bold text-slate-900 mb-4">Orders</h1>

      {/* Status Tabs */}
      <div className="flex gap-1 mb-4 overflow-x-auto pb-1">
        {TABS.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-1.5 rounded-full text-sm font-medium capitalize whitespace-nowrap transition-colors ${
              activeTab === tab
                ? 'bg-amethyst-500 text-white'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {!orders ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-100 rounded-lg h-20 animate-pulse" />
          ))}
        </div>
      ) : filteredOrders.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <svg className="w-12 h-12 mx-auto mb-3 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <p className="text-lg font-medium">No orders yet</p>
          <p className="text-sm mt-1">Share your storefront to start selling.</p>
        </div>
      ) : isMobile ? (
        /* Mobile: Card layout */
        <div className="space-y-3">
          {filteredOrders.map((order) => (
            <div key={order._id} className="bg-white rounded-lg shadow-sm p-4">
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-mono text-xs text-slate-500">#{shortOrderId(order._id)}</p>
                  <p className="text-sm font-medium text-slate-900 mt-0.5">{order.customerName}</p>
                </div>
                <OrderStatusBadge status={order.status} />
              </div>
              <div className="mt-2 text-sm text-slate-600">
                <p>Qty: {order.quantity} &middot; {formatMVR(order.totalAmount)}</p>
                <p className="text-xs text-slate-400 mt-0.5">
                  {order.deliveryLocation} &middot; {formatMaldivesTime(order.createdAt)}
                </p>
                {order.swipeReference && (
                  <p className="font-mono text-xs text-slate-400 mt-0.5">
                    Ref: {order.swipeReference}
                  </p>
                )}
              </div>
              <div className="mt-3 flex gap-2">
                {order.status === 'paid' && (
                  <button
                    onClick={() => handleFulfill(order._id, 'shipped')}
                    className="flex-1 bg-amethyst-500 text-white text-sm py-1.5 rounded-lg font-medium hover:bg-amethyst-600 transition-colors"
                  >
                    Mark Shipped
                  </button>
                )}
                {order.status === 'shipped' && (
                  <button
                    onClick={() => handleFulfill(order._id, 'delivered')}
                    className="flex-1 bg-amethyst-500 text-white text-sm py-1.5 rounded-lg font-medium hover:bg-amethyst-600 transition-colors"
                  >
                    Mark Delivered
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* Desktop: Table layout */
        <div className="bg-white rounded-lg shadow-sm overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-200 text-left text-slate-500 bg-slate-50">
                <th className="px-4 py-3 font-medium">Order</th>
                <th className="px-4 py-3 font-medium">Buyer</th>
                <th className="px-4 py-3 font-medium">Qty</th>
                <th className="px-4 py-3 font-medium">Amount</th>
                <th className="px-4 py-3 font-medium">Payment</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Swipe Ref</th>
                <th className="px-4 py-3 font-medium">Location</th>
                <th className="px-4 py-3 font-medium">Created</th>
                <th className="px-4 py-3 font-medium">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filteredOrders.map((order) => (
                <tr key={order._id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-mono text-xs text-slate-600">
                    #{shortOrderId(order._id)}
                  </td>
                  <td className="px-4 py-3 text-slate-900">{order.customerName}</td>
                  <td className="px-4 py-3 text-slate-600">{order.quantity}</td>
                  <td className="px-4 py-3 font-mono text-amethyst-700">{formatMVR(order.totalAmount)}</td>
                  <td className="px-4 py-3">
                    <OrderStatusBadge status={order.paymentStatus} />
                  </td>
                  <td className="px-4 py-3">
                    <OrderStatusBadge status={order.status} />
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-500">
                    {order.swipeReference ?? '-'}
                  </td>
                  <td className="px-4 py-3 text-slate-600 text-xs">{order.deliveryLocation}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">
                    {formatMaldivesTime(order.createdAt)}
                  </td>
                  <td className="px-4 py-3">
                    {order.status === 'paid' && (
                      <button
                        onClick={() => handleFulfill(order._id, 'shipped')}
                        className="bg-amethyst-500 text-white text-xs px-3 py-1 rounded-lg font-medium hover:bg-amethyst-600 transition-colors"
                      >
                        Ship
                      </button>
                    )}
                    {order.status === 'shipped' && (
                      <button
                        onClick={() => handleFulfill(order._id, 'delivered')}
                        className="bg-amethyst-500 text-white text-xs px-3 py-1 rounded-lg font-medium hover:bg-amethyst-600 transition-colors"
                      >
                        Deliver
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
