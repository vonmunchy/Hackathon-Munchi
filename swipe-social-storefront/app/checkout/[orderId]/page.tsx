"use client";

import { use, useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { useIsMobile } from "@/lib/use-device";
import { usePaymentStatus } from "@/lib/use-payment-status";
import { MvrAmount } from "@/components/shared/mvr-amount";

export default function CheckoutPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = use(params);
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const router = useRouter();
  const isMobile = useIsMobile();
  const [demoLoading, setDemoLoading] = useState<string | null>(null);

  const order = useQuery(api.orders.getByIdWithToken, {
    orderId: orderId as Id<"orders">,
    accessToken: token,
  });

  const { status: paymentStatus } = usePaymentStatus({
    orderId,
    accessToken: token,
    swipePaymentId: order?.swipePaymentId ?? null,
    enabled: !!order?.swipePaymentId && order?.paymentStatus === "pending",
  });

  // Auto-redirect on payment completion
  useEffect(() => {
    if (paymentStatus === "COMPLETED" || order?.paymentStatus === "paid") {
      const timer = setTimeout(() => {
        router.push(`/success/${orderId}?token=${token}`);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [paymentStatus, order?.paymentStatus, orderId, token, router]);

  // Loading state
  if (order === undefined || isMobile === null) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-amethyst-500 border-t-transparent" />
      </div>
    );
  }

  // Invalid token / not found
  if (order === null) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <h1 className="font-display text-xl font-semibold text-slate-800">
          Order not found
        </h1>
        <p className="text-slate-500">
          This order link may be invalid or expired.
        </p>
      </div>
    );
  }

  const isCompleted = paymentStatus === "COMPLETED" || order.paymentStatus === "paid";
  const isExpired = paymentStatus === "EXPIRED";
  const isCancelled = paymentStatus === "CANCELLED";
  const isTerminal = isExpired || isCancelled;
  const showDemoControls = process.env.NEXT_PUBLIC_SHOW_DEMO_CONTROLS === "true";

  async function handleSimulate(action: string) {
    if (!order?.swipeShortCode) return;
    setDemoLoading(action);
    try {
      await fetch("/api/swipe/payments/simulate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ shortCode: order.swipeShortCode, action }),
      });
    } catch {
      // ignore
    } finally {
      setDemoLoading(null);
    }
  }

  async function handleRetry() {
    // Create a new order with same delivery data
    try {
      const res = await fetch("/api/checkout/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          productId: order!.productId,
          variantId: order!.variantId,
          quantity: order!.quantity,
          customerName: order!.customerName,
          customerPhone: order!.customerPhone,
          deliveryLocation: order!.deliveryLocation,
          deliveryAddress: order!.deliveryAddress,
          deliveryTimePreference: order!.deliveryTimePreference,
        }),
      });

      const result = await res.json();
      if (res.ok) {
        router.push(`/checkout/${result.orderId}?token=${result.accessToken}`);
      }
    } catch {
      // ignore
    }
  }

  // QR code image
  const qrImage = order.swipeQrData ? (
    <img
      src={`data:image/png;base64,${order.swipeQrData}`}
      alt="Payment QR Code"
      className="mx-auto rounded-lg"
      width={isMobile ? 200 : 280}
      height={isMobile ? 200 : 280}
    />
  ) : (
    <div className="mx-auto flex h-48 w-48 items-center justify-center rounded-lg bg-slate-100 text-slate-400 text-sm">
      Generating QR...
    </div>
  );

  // Status indicator
  const statusIndicator = isCompleted ? (
    <div className="flex items-center justify-center gap-2 text-success">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-6 w-6"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={2.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
        />
      </svg>
      <span className="font-display font-semibold">Payment confirmed!</span>
    </div>
  ) : isTerminal ? (
    <div className="flex flex-col items-center gap-3">
      <div className="flex items-center gap-2 text-error">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
          />
        </svg>
        <span className="font-display font-semibold">
          {isExpired ? "Payment expired" : "Payment cancelled"}
        </span>
      </div>
      <button
        onClick={handleRetry}
        className="rounded-xl bg-amethyst-500 px-6 py-2.5 text-sm font-semibold text-white hover:bg-amethyst-600 transition-colors"
      >
        Try Again
      </button>
    </div>
  ) : (
    <div className="flex items-center justify-center gap-2 text-amethyst-600">
      <span className="relative flex h-3 w-3">
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-amethyst-400 opacity-75" />
        <span className="relative inline-flex h-3 w-3 rounded-full bg-amethyst-500" />
      </span>
      <span className="text-sm font-medium">Waiting for payment...</span>
    </div>
  );

  // Order summary
  const orderSummary = (
    <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
      <h3 className="text-sm font-medium text-slate-500 uppercase tracking-wide mb-3">
        Order Summary
      </h3>
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-slate-600">Amount</span>
          <MvrAmount amount={order.totalAmount} className="font-semibold text-slate-800" />
        </div>
        <div className="flex justify-between">
          <span className="text-slate-600">Quantity</span>
          <span className="text-slate-800">{order.quantity}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-slate-600">Delivery</span>
          <span className="text-slate-800">{order.deliveryLocation}</span>
        </div>
      </div>
    </div>
  );

  // Demo controls
  const demoControlsSection = showDemoControls && !isCompleted && !isTerminal && (
    <div className="mt-4 rounded-xl border border-dashed border-slate-300 bg-slate-50 p-4">
      <p className="mb-3 text-xs font-medium text-slate-500 uppercase tracking-wide">
        Demo Controls
      </p>
      <div className="flex flex-wrap gap-2">
        {[
          { action: "complete", label: "Simulate Payment", color: "bg-success" },
          { action: "expire", label: "Simulate Expiry", color: "bg-warning" },
          { action: "cancel", label: "Simulate Cancel", color: "bg-error" },
        ].map(({ action, label, color }) => (
          <button
            key={action}
            onClick={() => handleSimulate(action)}
            disabled={demoLoading !== null}
            className={`${color} rounded-lg px-3 py-1.5 text-xs font-medium text-white transition-opacity disabled:opacity-50`}
          >
            {demoLoading === action ? "..." : label}
          </button>
        ))}
      </div>
    </div>
  );

  // Pay with Swipe button
  const payButton = order.swipePaymentUrl && !isCompleted && !isTerminal && (
    <a
      href={order.swipePaymentUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="flex w-full items-center justify-center gap-2 rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600"
    >
      Pay with Swipe
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
          d="M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25"
        />
      </svg>
    </a>
  );

  if (isMobile) {
    return (
      <div className="mx-auto max-w-lg px-4 py-6">
        <h1 className="mb-6 text-center font-display text-xl font-bold text-slate-900">
          Complete Payment
        </h1>

        {/* Status */}
        <div className="mb-6">{statusIndicator}</div>

        {/* Pay button prominent on mobile */}
        {payButton && <div className="mb-5">{payButton}</div>}

        {/* QR */}
        {!isCompleted && !isTerminal && (
          <div className="mb-5 text-center">
            <p className="mb-3 text-sm text-slate-500">
              Or scan QR code to pay
            </p>
            {qrImage}
          </div>
        )}

        {/* Order summary */}
        {orderSummary}

        {/* Demo controls */}
        {demoControlsSection}
      </div>
    );
  }

  // Desktop two-column layout
  return (
    <div className="mx-auto max-w-4xl px-6 py-10">
      <h1 className="mb-8 text-center font-display text-2xl font-bold text-slate-900">
        Complete Payment
      </h1>

      <div className="grid grid-cols-2 gap-8">
        {/* Left: QR + Pay */}
        <div className="flex flex-col items-center gap-5">
          {!isCompleted && !isTerminal && (
            <>
              {qrImage}
              {payButton}
            </>
          )}

          {/* Status */}
          <div className="mt-2">{statusIndicator}</div>

          {/* Demo controls */}
          {demoControlsSection}
        </div>

        {/* Right: Order summary */}
        <div>{orderSummary}</div>
      </div>
    </div>
  );
}
