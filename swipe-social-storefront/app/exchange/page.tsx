"use client";

import Link from "next/link";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import { formatMVR } from "@/lib/format";

export default function ExchangeBrowsePage() {
  const listings = useQuery(api.exchange.getActiveListings);

  // Loading state
  if (listings === undefined) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6">
        <div className="mb-6">
          <h1 className="font-display text-xl font-bold text-slate-900 md:text-2xl">
            Buy USDT
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Secure P2P exchange powered by Swipe
          </p>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="animate-pulse rounded-xl border border-slate-200 bg-white p-5"
            >
              <div className="mb-3 h-4 w-24 rounded bg-slate-200" />
              <div className="mb-2 h-6 w-32 rounded bg-slate-200" />
              <div className="mb-2 h-4 w-40 rounded bg-slate-200" />
              <div className="mb-4 h-4 w-28 rounded bg-slate-200" />
              <div className="h-10 w-full rounded-xl bg-slate-200" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Empty state
  if (listings.length === 0) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6">
        <div className="mb-6">
          <h1 className="font-display text-xl font-bold text-slate-900 md:text-2xl">
            Buy USDT
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Secure P2P exchange powered by Swipe
          </p>
        </div>
        <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-100">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-8 w-8 text-slate-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
              />
            </svg>
          </div>
          <h2 className="font-display text-lg font-semibold text-slate-800">
            No listings available
          </h2>
          <p className="max-w-xs text-sm text-slate-500">
            Check back soon — sellers are adding USDT listings
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-6">
      <div className="mb-6">
        <h1 className="font-display text-xl font-bold text-slate-900 md:text-2xl">
          Buy USDT
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          Secure P2P exchange powered by Swipe
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {listings.map((listing) => {
          const totalMvr = listing.availableBalance * listing.rate;
          return (
            <div
              key={listing._id}
              className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
            >
              {/* Seller name */}
              <p className="text-sm font-medium text-slate-500">
                {listing.storeName}
              </p>

              {/* USDT available */}
              <p className="mt-2 font-display text-lg font-bold text-slate-900">
                {listing.availableBalance.toLocaleString()} USDT
              </p>

              {/* Rate */}
              <p className="mt-1 text-sm text-slate-600">
                Rate:{" "}
                <span className="font-semibold text-slate-800">
                  {formatMVR(listing.rate)}
                </span>
                <span className="text-slate-400"> / USDT</span>
              </p>

              {/* Total MVR */}
              <p className="mt-0.5 text-sm text-slate-500">
                Total: {formatMVR(totalMvr)}
              </p>

              {/* Buy button */}
              <Link
                href={`/exchange/buy/${listing._id}`}
                className="mt-4 flex w-full items-center justify-center rounded-xl bg-amethyst-500 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-amethyst-600"
              >
                Buy
              </Link>
            </div>
          );
        })}
      </div>
    </div>
  );
}
