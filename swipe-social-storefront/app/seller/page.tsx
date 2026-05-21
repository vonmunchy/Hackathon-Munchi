'use client'

import { useState, useEffect } from 'react'
import { useQuery } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useSessionToken } from '@/lib/use-session'
import { formatMVR } from '@/lib/format'
import { DashboardCard } from '@/components/seller/dashboard-card'
import Link from 'next/link'

export default function SellerDashboardPage() {
  const sessionToken = useSessionToken()
  const [storeSlug, setStoreSlug] = useState<string | null>(null)

  useEffect(() => {
    setStoreSlug(localStorage.getItem('storeSlug'))
  }, [])

  const store = useQuery(api.stores.getBySlug, storeSlug ? { slug: storeSlug } : 'skip')
  const stats = useQuery(
    api.orders.getDashboardStats,
    sessionToken ? { sessionToken } : 'skip',
  )
  const recentOrders = useQuery(
    api.orders.listByStore,
    sessionToken ? { sessionToken } : 'skip',
  )

  if (!storeSlug) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="w-8 h-8 border-2 border-amethyst-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  const storeName = store?.name ?? localStorage.getItem('storeName') ?? 'Store'
  const sellerType = store?.sellerType ?? 'marketplace'
  const showMarketplace = sellerType === 'marketplace' || sellerType === 'both'
  const showCrypto = sellerType === 'crypto' || sellerType === 'both'

  return (
    <div className="p-4 md:p-6 max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-display font-bold text-slate-900">Dashboard</h1>
        <p className="text-slate-500 text-sm">{storeName}</p>
      </div>

      {/* Marketplace Section */}
      {showMarketplace && (
        <>
          {sellerType === 'both' && (
            <h2 className="text-xl font-display font-semibold text-slate-800">Marketplace</h2>
          )}

          {/* Stats Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
            <DashboardCard
              title="Total Sales"
              value={stats ? formatMVR(stats.totalSalesAllTime) : '---'}
              subtitle={stats?.totalSalesToday ? `Today: ${formatMVR(stats.totalSalesToday)}` : undefined}
              accentColor="text-amethyst-700"
            />
            <DashboardCard
              title="Paid Orders"
              value={stats ? String(stats.paidCount) : '---'}
              subtitle={stats?.paidCount ? 'orders confirmed' : undefined}
              accentColor="text-green-600"
            />
            <DashboardCard
              title="Awaiting Shipment"
              value={stats ? String(stats.awaitingShipmentCount) : '---'}
              subtitle={stats?.awaitingShipmentCount ? 'need fulfillment' : undefined}
              accentColor="text-ruby-600"
            />
            <DashboardCard
              title="Low Stock"
              value={stats ? String(stats.lowStockCount) : '---'}
              subtitle={stats?.lowStockCount ? 'variants low' : undefined}
              accentColor="text-yellow-600"
            />
          </div>

          {/* Recent Orders */}
          <div>
            <h2 className="text-lg font-display font-semibold text-slate-800 mb-3">
              Recent Orders
            </h2>
            {!recentOrders ? (
              <div className="space-y-2">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="bg-slate-100 rounded-lg h-14 animate-pulse" />
                ))}
              </div>
            ) : recentOrders.length === 0 ? (
              <div className="text-center py-8 text-slate-400">
                <p className="text-sm">No orders yet</p>
                <p className="text-xs mt-1">Orders will appear here after your first sale</p>
              </div>
            ) : (
              <div className="space-y-2">
                {recentOrders.slice(0, 10).map((order) => (
                  <div key={order._id} className="flex items-center justify-between bg-white rounded-lg border border-slate-100 px-4 py-3">
                    <div>
                      <p className="text-sm font-medium text-slate-800">{order.customerName}</p>
                      <p className="text-xs text-slate-500">{order.deliveryLocation} &middot; {new Date(order.createdAt).toLocaleDateString()}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm font-mono font-semibold text-amethyst-700">{formatMVR(order.totalAmount)}</p>
                      <span className={`inline-block text-xs px-2 py-0.5 rounded-full font-medium ${
                        order.paymentStatus === 'paid'
                          ? 'bg-green-50 text-green-700'
                          : 'bg-slate-100 text-slate-500'
                      }`}>
                        {order.paymentStatus === 'paid' ? 'Paid' : 'Pending'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </>
      )}

      {/* Divider between sections */}
      {showMarketplace && showCrypto && (
        <div className="border-t border-slate-200 my-6" />
      )}

      {/* Crypto Exchange Section */}
      {showCrypto && (
        <>
          {sellerType === 'both' && (
            <h2 className="text-xl font-display font-semibold text-slate-800">Crypto Exchange</h2>
          )}

          {/* Crypto Stats Grid */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3 md:gap-4">
            <DashboardCard
              title="Active Listings"
              value="---"
              subtitle="View Exchange"
              accentColor="text-amethyst-700"
            />
            <DashboardCard
              title="USDT Escrowed"
              value="---"
              subtitle="in active listings"
              accentColor="text-green-600"
            />
            <DashboardCard
              title="Completed Trades"
              value="---"
              subtitle="all time"
              accentColor="text-blue-600"
            />
          </div>

          {/* Quick Actions */}
          <div className="flex flex-col sm:flex-row gap-3">
            <Link
              href="/seller/exchange/new"
              className="flex-1 flex items-center justify-center gap-2 bg-amethyst-600 hover:bg-amethyst-700 text-white font-medium rounded-lg px-4 py-3 transition-colors"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
              </svg>
              Create Listing
            </Link>
            <Link
              href="/seller/exchange"
              className="flex-1 flex items-center justify-center gap-2 bg-white border border-slate-200 hover:border-amethyst-300 text-slate-700 font-medium rounded-lg px-4 py-3 transition-colors"
            >
              View Exchange Dashboard &rarr;
            </Link>
          </div>
        </>
      )}
    </div>
  )
}
