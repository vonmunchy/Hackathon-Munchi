"use client";

import { useState, useEffect, useRef, use } from "react";
import { useQuery } from "convex/react";
import { useRouter } from "next/navigation";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { formatMVR } from "@/lib/format";

type Step = 1 | 2 | 3 | 4;

export default function ExchangeBuyPage({
  params,
}: {
  params: Promise<{ listingId: string }>;
}) {
  const { listingId } = use(params);
  const router = useRouter();

  const listing = useQuery(api.exchange.getListingDetails, {
    listingId: listingId as Id<"exchangeListings">,
  });

  // Form state
  const [step, setStep] = useState<Step>(1);
  const [usdtAmount, setUsdtAmount] = useState("");
  const [buyerWallet, setBuyerWallet] = useState("");
  const [walletError, setWalletError] = useState("");

  // Purchase state
  const [purchaseLoading, setPurchaseLoading] = useState(false);
  const [purchaseError, setPurchaseError] = useState("");
  const [reservationId, setReservationId] = useState<string | null>(null);
  const [paymentId, setPaymentId] = useState<string | null>(null);
  const [paymentUrl, setPaymentUrl] = useState<string | null>(null);

  // Completion state
  const [txHash, setTxHash] = useState<string | null>(null);
  const [completedAmount, setCompletedAmount] = useState<number>(0);
  const [pollStatus, setPollStatus] = useState<string>("pending");
  const [pollTimedOut, setPollTimedOut] = useState(false);

  // Polling ref
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const pollStartRef = useRef<number>(0);

  // Auto-skip step 1 if partialAllowed is false
  useEffect(() => {
    if (listing && !listing.partialAllowed && step === 1) {
      setUsdtAmount(String(listing.availableBalance));
      setStep(2);
    }
  }, [listing, step]);

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
      }
    };
  }, []);

  // Start polling when we enter step 4
  useEffect(() => {
    if (step !== 4 || !reservationId || !paymentId) return;

    pollStartRef.current = Date.now();

    pollIntervalRef.current = setInterval(async () => {
      // Timeout after 5 minutes
      if (Date.now() - pollStartRef.current > 5 * 60 * 1000) {
        if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
        setPollTimedOut(true);
        return;
      }

      try {
        const res = await fetch(
          `/api/exchange/purchase/${reservationId}/status?paymentId=${paymentId}`
        );
        if (res.ok) {
          const data = await res.json();
          if (data.status === "completed") {
            setPollStatus("completed");
            setTxHash(data.txHash ?? null);
            setCompletedAmount(data.usdtAmount ?? parseFloat(usdtAmount));
            if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
          } else if (data.status === "failed" || data.status === "expired") {
            setPollStatus("failed");
            if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
          }
        }
      } catch {
        // Continue polling on error
      }
    }, 2000);

    return () => {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
    };
  }, [step, reservationId, paymentId, usdtAmount]);

  // Loading state
  if (listing === undefined) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-amethyst-500 border-t-transparent" />
      </div>
    );
  }

  // Not found
  if (listing === null) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <h1 className="font-display text-xl font-semibold text-slate-800">
          Listing not found
        </h1>
        <p className="text-slate-500">
          This listing may have been removed or sold out.
        </p>
        <button
          onClick={() => router.push("/exchange")}
          className="mt-3 rounded-xl bg-amethyst-500 px-5 py-2.5 text-sm font-semibold text-white hover:bg-amethyst-600"
        >
          Back to Exchange
        </button>
      </div>
    );
  }

  const parsedAmount = parseFloat(usdtAmount) || 0;
  const mvrCost = parsedAmount * listing.rate;
  const isValidAmount = parsedAmount >= 1 && parsedAmount <= listing.availableBalance;

  // TRC20 wallet validation
  const walletRegex = /^T[A-Za-z0-9]{33}$/;
  const isValidWallet = walletRegex.test(buyerWallet);

  function handleWalletChange(value: string) {
    setBuyerWallet(value);
    if (value && !walletRegex.test(value)) {
      setWalletError("Invalid TRC20 address format");
    } else {
      setWalletError("");
    }
  }

  async function handlePaste() {
    try {
      const text = await navigator.clipboard.readText();
      handleWalletChange(text.trim());
    } catch {
      // Clipboard access denied
    }
  }

  async function handlePurchase() {
    setPurchaseLoading(true);
    setPurchaseError("");
    try {
      const res = await fetch("/api/exchange/purchase", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          listingId,
          usdtAmount: parsedAmount,
          buyerWallet,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setPurchaseError(data.error || "Failed to create purchase");
        return;
      }
      setReservationId(data.reservationId);
      setPaymentId(data.paymentId);
      setPaymentUrl(data.paymentUrl || null);
      setStep(4);
    } catch {
      setPurchaseError("Something went wrong. Please try again.");
    } finally {
      setPurchaseLoading(false);
    }
  }

  // Step indicator
  const stepIndicator = (
    <div className="mb-6 flex items-center justify-center gap-2">
      {[1, 2, 3, 4].map((s) => (
        <div
          key={s}
          className={`h-2 w-2 rounded-full transition-colors ${
            s === step
              ? "bg-amethyst-500"
              : s < step
                ? "bg-amethyst-300"
                : "bg-slate-200"
          }`}
        />
      ))}
    </div>
  );

  // Listing info header
  const listingInfo = (
    <div className="mb-6 rounded-xl border border-slate-200 bg-slate-50 p-4">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-slate-500">Seller</p>
          <p className="font-medium text-slate-800">{listing.storeName}</p>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Rate</p>
          <p className="font-semibold text-amethyst-600">
            {formatMVR(listing.rate)}/USDT
          </p>
        </div>
      </div>
      <div className="mt-3 border-t border-slate-200 pt-3">
        <p className="text-sm text-slate-500">Available</p>
        <p className="font-display text-lg font-bold text-slate-900">
          {listing.availableBalance.toLocaleString()} USDT
        </p>
      </div>
    </div>
  );

  // Step 1: Amount
  const step1Content = (
    <div className="space-y-4 transition-opacity duration-200">
      <h2 className="font-display text-lg font-semibold text-slate-900">
        How much USDT?
      </h2>
      <div className="relative">
        <input
          type="number"
          inputMode="decimal"
          min={1}
          max={listing.availableBalance}
          value={usdtAmount}
          onChange={(e) => setUsdtAmount(e.target.value)}
          placeholder="Enter USDT amount"
          className="w-full rounded-xl border border-slate-300 px-4 py-3 text-lg font-medium text-slate-800 placeholder:text-slate-400 focus:border-amethyst-500 focus:outline-none focus:ring-2 focus:ring-amethyst-500/20"
        />
        <button
          onClick={() => setUsdtAmount(String(listing.availableBalance))}
          className="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg bg-amethyst-100 px-3 py-1 text-xs font-semibold text-amethyst-600 hover:bg-amethyst-200"
        >
          MAX
        </button>
      </div>
      {parsedAmount > 0 && (
        <div className="rounded-lg bg-amethyst-50 px-4 py-3">
          <p className="text-sm text-slate-600">
            You will pay{" "}
            <span className="font-semibold text-amethyst-700">
              {formatMVR(mvrCost)}
            </span>
          </p>
        </div>
      )}
      <button
        onClick={() => setStep(2)}
        disabled={!isValidAmount}
        className="w-full rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:cursor-not-allowed disabled:bg-slate-300"
      >
        Continue
      </button>
    </div>
  );

  // Step 2: Wallet Address
  const step2Content = (
    <div className="space-y-4 transition-opacity duration-200">
      <h2 className="font-display text-lg font-semibold text-slate-900">
        Your TRC20 wallet
      </h2>
      <p className="text-sm text-slate-500">
        Your USDT will be sent to this address
      </p>
      <div className="relative">
        <input
          type="text"
          value={buyerWallet}
          onChange={(e) => handleWalletChange(e.target.value)}
          placeholder="T..."
          className={`w-full rounded-xl border px-4 py-3 font-mono text-sm text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 ${
            walletError
              ? "border-red-400 focus:border-red-500 focus:ring-red-500/20"
              : "border-slate-300 focus:border-amethyst-500 focus:ring-amethyst-500/20"
          }`}
        />
        <button
          onClick={handlePaste}
          className="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600 hover:bg-slate-200"
        >
          Paste
        </button>
      </div>
      {walletError && (
        <p className="text-xs text-red-500">{walletError}</p>
      )}
      <p className="text-xs text-slate-400">
        TRC20 format — starts with T, 34 characters
      </p>
      <div className="flex gap-3">
        <button
          onClick={() => setStep(1)}
          className="rounded-xl border border-slate-300 px-5 py-3 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          Back
        </button>
        <button
          onClick={() => setStep(3)}
          disabled={!isValidWallet}
          className="flex-1 rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Continue
        </button>
      </div>
    </div>
  );

  // Step 3: Review & Pay
  const step3Content = (
    <div className="space-y-4 transition-opacity duration-200">
      <h2 className="font-display text-lg font-semibold text-slate-900">
        Review & Pay
      </h2>
      <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-4">
        <div className="flex justify-between">
          <span className="text-sm text-slate-500">USDT Amount</span>
          <span className="font-semibold text-slate-800">
            {parsedAmount.toLocaleString()} USDT
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-sm text-slate-500">Rate</span>
          <span className="text-sm text-slate-800">
            {formatMVR(listing.rate)} / USDT
          </span>
        </div>
        <div className="border-t border-slate-100 pt-3">
          <div className="flex justify-between">
            <span className="text-sm font-medium text-slate-700">
              Total MVR
            </span>
            <span className="font-display font-bold text-amethyst-600">
              {formatMVR(mvrCost)}
            </span>
          </div>
        </div>
        <div className="border-t border-slate-100 pt-3">
          <div className="flex justify-between">
            <span className="text-sm text-slate-500">Destination</span>
            <span className="font-mono text-xs text-slate-600">
              {buyerWallet.slice(0, 8)}...{buyerWallet.slice(-6)}
            </span>
          </div>
        </div>
      </div>

      {purchaseError && (
        <div className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
          {purchaseError}
        </div>
      )}

      <div className="flex gap-3">
        <button
          onClick={() => setStep(2)}
          className="rounded-xl border border-slate-300 px-5 py-3 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          Back
        </button>
        <button
          onClick={handlePurchase}
          disabled={purchaseLoading}
          className="flex-1 rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:opacity-60"
        >
          {purchaseLoading ? (
            <span className="flex items-center justify-center gap-2">
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              Processing...
            </span>
          ) : (
            `Pay ${formatMVR(mvrCost)} with Swipe`
          )}
        </button>
      </div>
    </div>
  );

  // Step 4: Payment & Confirmation
  const step4Content = (
    <div className="space-y-5 transition-opacity duration-200">
      {pollStatus === "completed" ? (
        // Success state
        <div className="flex flex-col items-center gap-4 py-4 text-center">
          <div className="relative flex h-16 w-16 items-center justify-center">
            <div className="absolute inset-0 animate-ping rounded-full bg-green-200 opacity-50" />
            <div className="relative flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-8 w-8 text-green-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M4.5 12.75l6 6 9-13.5"
                />
              </svg>
            </div>
          </div>
          <h2 className="font-display text-xl font-bold text-slate-900">
            USDT Transferred!
          </h2>
          {txHash && (
            <div className="w-full rounded-lg bg-slate-50 px-4 py-3">
              <p className="text-xs text-slate-500">Transaction Hash</p>
              <p className="mt-1 break-all font-mono text-xs text-slate-700">
                {txHash.slice(0, 20)}...{txHash.slice(-12)}
              </p>
            </div>
          )}
          <p className="text-sm text-slate-600">
            <span className="font-semibold">{completedAmount} USDT</span>{" "}
            sent to your wallet
          </p>
          <button
            onClick={() => router.push("/exchange")}
            className="mt-2 w-full rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white hover:bg-amethyst-600"
          >
            Back to Exchange
          </button>
        </div>
      ) : pollStatus === "failed" ? (
        // Failed state
        <div className="flex flex-col items-center gap-4 py-4 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-8 w-8 text-red-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 className="font-display text-lg font-semibold text-slate-900">
            Payment Failed
          </h2>
          <p className="text-sm text-slate-500">
            The payment could not be completed. Please try again.
          </p>
          <button
            onClick={() => {
              setPollStatus("pending");
              setPurchaseError("");
              setStep(3);
            }}
            className="mt-2 w-full rounded-xl bg-amethyst-500 py-3 text-base font-semibold text-white hover:bg-amethyst-600"
          >
            Try Again
          </button>
        </div>
      ) : pollTimedOut ? (
        // Timeout state
        <div className="flex flex-col items-center gap-4 py-4 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-amber-100">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-8 w-8 text-amber-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
              />
            </svg>
          </div>
          <h2 className="font-display text-lg font-semibold text-slate-900">
            Payment not detected
          </h2>
          <p className="text-sm text-slate-500">
            We could not confirm your payment within the time limit. If you
            completed the payment, please contact support.
          </p>
          <button
            onClick={() => router.push("/exchange")}
            className="mt-2 w-full rounded-xl border border-slate-300 py-3 text-sm font-medium text-slate-600 hover:bg-slate-50"
          >
            Back to Exchange
          </button>
        </div>
      ) : (
        // Waiting state
        <div className="flex flex-col items-center gap-5 py-4 text-center">
          <h2 className="font-display text-lg font-semibold text-slate-900">
            Complete Payment
          </h2>

          {paymentUrl && (
            <a
              href={paymentUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex w-full items-center justify-center gap-2 rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600"
            >
              Complete Payment
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
          )}

          <div className="flex items-center gap-2 text-amethyst-600">
            <span className="relative flex h-3 w-3">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-amethyst-400 opacity-75" />
              <span className="relative inline-flex h-3 w-3 rounded-full bg-amethyst-500" />
            </span>
            <span className="text-sm font-medium">
              Waiting for payment confirmation...
            </span>
          </div>

          <p className="text-xs text-slate-400">
            This page will update automatically once payment is received
          </p>
        </div>
      )}
    </div>
  );

  // Render current step content
  let currentContent;
  switch (step) {
    case 1:
      currentContent = step1Content;
      break;
    case 2:
      currentContent = step2Content;
      break;
    case 3:
      currentContent = step3Content;
      break;
    case 4:
      currentContent = step4Content;
      break;
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-6">
      {stepIndicator}
      {step < 4 && listingInfo}
      {currentContent}
    </div>
  );
}
