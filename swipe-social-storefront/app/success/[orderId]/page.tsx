"use client";

import { use, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { MvrAmount } from "@/components/shared/mvr-amount";
import { shortOrderId } from "@/lib/format";
import Link from "next/link";

export default function SuccessPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = use(params);
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const router = useRouter();

  const order = useQuery(api.orders.getByIdWithToken, {
    orderId: orderId as Id<"orders">,
    accessToken: token,
  });

  const orderDetails = useQuery(api.orders.getOrderWithDetails, {
    orderId: orderId as Id<"orders">,
  });

  // Get store info once we have the order's storeId
  const store = useQuery(
    api.stores.getById,
    order ? { storeId: order.storeId } : "skip",
  );

  // Redirect if not paid
  useEffect(() => {
    if (order && order.paymentStatus !== "paid") {
      router.replace(`/checkout/${orderId}?token=${token}`);
    }
  }, [order, orderId, token, router]);

  // Loading
  if (order === undefined || orderDetails === undefined) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-amethyst-500 border-t-transparent" />
      </div>
    );
  }

  // Not found
  if (order === null) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <h1 className="font-display text-xl font-semibold text-slate-800">
          Order not found
        </h1>
      </div>
    );
  }

  // Not yet paid — will redirect
  if (order.paymentStatus !== "paid") {
    return null;
  }

  const product = orderDetails?.product;
  const variant = orderDetails?.variant;
  const orderRef = shortOrderId(orderId);

  // Instagram DM link
  const igLink = store?.instagramUsername
    ? `https://ig.me/m/${store.instagramUsername}`
    : null;

  // Facebook Messenger link
  const fbPageUrl = store?.facebookPageUrl;
  const fbPageId = fbPageUrl
    ? fbPageUrl.replace(/.*facebook\.com\//, "").replace(/\/$/, "")
    : null;
  const fbLink = fbPageId
    ? `https://m.me/${fbPageId}`
    : null;

  // Back to store link
  const storeLink = store ? `/shop/${store.slug}` : "/";

  return (
    <div className="mx-auto max-w-lg px-4 py-10">
      {/* Animated checkmark */}
      <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-success/10">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-10 w-10 text-success"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="m4.5 12.75 6 6 9-13.5"
          />
        </svg>
      </div>

      <h1 className="text-center font-display text-2xl font-bold text-slate-900">
        Payment Confirmed!
      </h1>
      <p className="mt-2 text-center text-slate-500">Order #{orderRef}</p>

      {/* Order details card */}
      <div className="mt-6 rounded-xl border border-slate-200 bg-slate-50 p-5 space-y-3">
        {product && (
          <div className="flex justify-between">
            <span className="text-sm text-slate-600">Product</span>
            <span className="text-sm font-medium text-slate-800 text-right max-w-[60%]">
              {product.name}
            </span>
          </div>
        )}
        {variant && (
          <div className="flex justify-between">
            <span className="text-sm text-slate-600">Option</span>
            <span className="text-sm text-slate-800">
              {variant.variantName}
            </span>
          </div>
        )}
        <div className="flex justify-between">
          <span className="text-sm text-slate-600">Quantity</span>
          <span className="text-sm text-slate-800">{order.quantity}</span>
        </div>
        <div className="flex justify-between border-t border-slate-200 pt-3">
          <span className="text-sm font-medium text-slate-700">Total</span>
          <MvrAmount
            amount={order.totalAmount}
            className="font-semibold text-amethyst-600"
          />
        </div>
      </div>

      {/* Delivery details */}
      <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-5 space-y-2">
        <h3 className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-2">
          Delivery
        </h3>
        <p className="text-sm text-slate-800">{order.customerName}</p>
        <p className="text-sm text-slate-600">{order.deliveryLocation}</p>
        <p className="text-sm text-slate-600">{order.deliveryAddress}</p>
        {order.deliveryTimePreference && (
          <p className="text-sm text-slate-500 italic">
            {order.deliveryTimePreference}
          </p>
        )}
      </div>

      {/* Contact seller */}
      <div className="mt-6 space-y-3">
        <h3 className="text-center text-sm font-medium text-slate-600">
          Message the Seller
        </h3>
        <div className="flex gap-3">
          {/* Instagram DM */}
          {igLink && (
            <a
              href={igLink}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-purple-500 via-pink-500 to-orange-400 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90"
            >
              <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zM12 0C8.741 0 8.333.014 7.053.072 2.695.272.273 2.69.073 7.052.014 8.333 0 8.741 0 12c0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98C8.333 23.986 8.741 24 12 24c3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 1 0 0 12.324 6.162 6.162 0 0 0 0-12.324zM12 16a4 4 0 1 1 0-8 4 4 0 0 1 0 8zm6.406-11.845a1.44 1.44 0 1 0 0 2.881 1.44 1.44 0 0 0 0-2.881z" />
              </svg>
              Instagram
            </a>
          )}

          {/* Facebook Messenger */}
          {fbLink && (
            <a
              href={fbLink}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-[#0084FF] py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90"
            >
              <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0C5.373 0 0 4.974 0 11.111c0 3.497 1.745 6.616 4.472 8.652V24l4.086-2.242c1.09.301 2.246.464 3.442.464 6.627 0 12-4.974 12-11.111S18.627 0 12 0zm1.193 14.963-3.056-3.259-5.963 3.259L10.733 8.2l3.13 3.259L19.752 8.2l-6.559 6.763z" />
              </svg>
              Facebook
            </a>
          )}
        </div>
      </div>

      {/* Back to store */}
      <div className="mt-8 text-center">
        <Link
          href={storeLink}
          className="inline-flex items-center gap-1.5 text-sm font-medium text-amethyst-600 hover:text-amethyst-700 transition-colors"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
            />
          </svg>
          Back to Store
        </Link>
      </div>
    </div>
  );
}
