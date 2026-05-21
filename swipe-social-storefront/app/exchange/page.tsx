"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import { formatMVR } from "@/lib/format";

type Tab = "buy" | "sell";

export default function ExchangeBrowsePage() {
  const [activeTab, setActiveTab] = useState<Tab>("buy");
  const listings = useQuery(api.exchange.getActiveListings);

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      {/* Header */}
      <div className="mb-5">
        <h1 className="font-display text-xl font-bold text-slate-900 md:text-2xl">
          P2P Trading
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          Buy and sell USDT with MVR — powered by Swipe
        </p>
      </div>

      {/* Buy / Sell Tab Switcher */}
      <div className="mb-4 flex gap-1 rounded-lg bg-slate-100 p-1 w-fit">
        <button
          onClick={() => setActiveTab("buy")}
          className={`rounded-md px-6 py-2 text-sm font-semibold transition-colors ${
            activeTab === "buy"
              ? "bg-success text-white shadow-sm"
              : "text-slate-500 hover:text-slate-700"
          }`}
        >
          Buy
        </button>
        <button
          onClick={() => setActiveTab("sell")}
          className={`rounded-md px-6 py-2 text-sm font-semibold transition-colors ${
            activeTab === "sell"
              ? "bg-ruby-500 text-white shadow-sm"
              : "text-slate-500 hover:text-slate-700"
          }`}
        >
          Sell
        </button>
      </div>

      {/* Tab Content */}
      {activeTab === "buy" ? (
        <BuyTabContent listings={listings} />
      ) : (
        <SellTabContent />
      )}
    </div>
  );
}

function BuyTabContent({
  listings,
}: {
  listings:
    | Array<{
        _id: string;
        storeName: string;
        rate: number;
        availableBalance: number;
        usdtAmount: number;
      }>
    | undefined;
}) {
  // Loading skeleton — table rows
  if (listings === undefined) {
    return (
      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <table className="w-full">
          <thead>
            <tr className="border-b border-slate-100 bg-slate-50/80">
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                Advertiser
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                Price
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 sm:table-cell">
                Available
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 md:table-cell">
                Limits
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-slate-500">
                Trade
              </th>
            </tr>
          </thead>
          <tbody>
            {[1, 2, 3, 4, 5].map((i) => (
              <tr key={i} className="border-b border-slate-50">
                <td className="px-4 py-4">
                  <div className="h-4 w-24 animate-pulse rounded bg-slate-200" />
                </td>
                <td className="px-4 py-4">
                  <div className="h-4 w-20 animate-pulse rounded bg-slate-200" />
                </td>
                <td className="hidden px-4 py-4 sm:table-cell">
                  <div className="h-4 w-16 animate-pulse rounded bg-slate-200" />
                </td>
                <td className="hidden px-4 py-4 md:table-cell">
                  <div className="h-4 w-24 animate-pulse rounded bg-slate-200" />
                </td>
                <td className="px-4 py-4 text-right">
                  <div className="ml-auto h-8 w-20 animate-pulse rounded-lg bg-slate-200" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  // Empty state
  if (listings.length === 0) {
    return (
      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <table className="w-full">
          <thead>
            <tr className="border-b border-slate-100 bg-slate-50/80">
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                Advertiser
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                Price
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 sm:table-cell">
                Available
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 md:table-cell">
                Limits
              </th>
              <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-slate-500">
                Trade
              </th>
            </tr>
          </thead>
        </table>
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-slate-100">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-7 w-7 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
            </svg>
          </div>
          <p className="mt-3 font-display text-sm font-semibold text-slate-700">
            No listings available
          </p>
          <p className="mt-1 text-xs text-slate-500">
            Check back soon — sellers are adding USDT listings
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-slate-100 bg-slate-50/80">
              <th className="px-3 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 md:px-4 md:py-3">
                Advertiser
              </th>
              <th className="px-3 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 md:px-4 md:py-3">
                Price
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 sm:table-cell">
                Available
              </th>
              <th className="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 md:table-cell">
                Limits
              </th>
              <th className="px-3 py-2.5 text-right text-xs font-semibold uppercase tracking-wider text-slate-500 md:px-4 md:py-3">
                Trade
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {listings.map((listing) => {
              const minMvr = formatMVR(listing.rate);
              const maxMvr = formatMVR(listing.availableBalance * listing.rate);
              return (
                <tr
                  key={listing._id}
                  className="transition-colors hover:bg-slate-50/50"
                >
                  {/* Advertiser */}
                  <td className="px-3 py-2.5 md:px-4 md:py-3.5">
                    <div className="flex items-center gap-2">
                      <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-amethyst-100 text-xs font-bold text-amethyst-600">
                        {listing.storeName.charAt(0).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <span className="block max-w-[100px] truncate text-sm font-medium text-slate-800 sm:max-w-[120px]">
                          {listing.storeName}
                        </span>
                        <span className="block text-xs text-slate-400 sm:hidden">
                          {listing.availableBalance.toLocaleString()} USDT
                        </span>
                      </div>
                    </div>
                  </td>

                  {/* Price */}
                  <td className="px-3 py-2.5 md:px-4 md:py-3.5">
                    <span className="font-mono text-sm font-semibold text-slate-900">
                      {formatMVR(listing.rate)}
                    </span>
                    <span className="ml-1 text-xs text-slate-400">MVR</span>
                  </td>

                  {/* Available — hidden on mobile */}
                  <td className="hidden px-4 py-3.5 sm:table-cell">
                    <span className="font-mono text-sm text-slate-700">
                      {listing.availableBalance.toLocaleString()}
                    </span>
                    <span className="ml-1 text-xs text-slate-400">USDT</span>
                  </td>

                  {/* Limits — hidden on mobile/tablet */}
                  <td className="hidden px-4 py-3.5 md:table-cell">
                    <span className="text-xs text-slate-500">
                      {minMvr} – {maxMvr}
                    </span>
                  </td>

                  {/* Action */}
                  <td className="px-3 py-2.5 text-right md:px-4 md:py-3.5">
                    <Link
                      href={`/exchange/buy/${listing._id}`}
                      className="inline-flex items-center rounded-lg bg-success px-4 py-2 text-xs font-bold text-white transition-colors hover:bg-emerald-600"
                    >
                      Buy USDT
                    </Link>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function SellTabContent() {
  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div className="flex flex-col items-center justify-center px-6 py-16 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-ruby-50">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-ruby-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 0 0-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 0 1-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 0 0 3 15h-.75M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm3 0h.008v.008H18V10.5Zm-12 0h.008v.008H6V10.5Z" />
          </svg>
        </div>
        <h2 className="mt-4 font-display text-lg font-semibold text-slate-800">
          Want to sell USDT?
        </h2>
        <p className="mt-2 max-w-sm text-sm text-slate-500">
          Create a seller account and list your USDT for sale. Set your own rate and trade limits.
        </p>
        <Link
          href="/seller/exchange/new"
          className="mt-6 inline-flex items-center rounded-xl bg-ruby-500 px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-ruby-600"
        >
          Start Selling
        </Link>
      </div>
    </div>
  );
}
