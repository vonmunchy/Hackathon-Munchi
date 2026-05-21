"use client";

import { useState, use, useEffect } from "react";
import { useQuery } from "convex/react";
import { useRouter, useSearchParams } from "next/navigation";
import Image from "next/image";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { useIsMobile } from "@/lib/use-device";
import { formatMVR } from "@/lib/format";
import { MvrAmount } from "@/components/shared/mvr-amount";
import { VariantSelector } from "@/components/buyer/variant-selector";
import { QuantitySelector } from "@/components/buyer/quantity-selector";
import { DeliveryForm, type DeliveryData } from "@/components/buyer/delivery-form";

export default function ProductDetailPage({
  params,
}: {
  params: Promise<{ storeSlug: string; productId: string }>;
}) {
  const { storeSlug, productId } = use(params);
  const isMobile = useIsMobile();

  const data = useQuery(api.products.getWithVariants, {
    productId: productId as Id<"products">,
  });

  const store = useQuery(api.stores.getBySlug, { slug: storeSlug });

  const router = useRouter();
  const searchParams = useSearchParams();
  const [selectedVariantId, setSelectedVariantId] =
    useState<Id<"productVariants"> | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [copied, setCopied] = useState(false);
  const [showDeliveryForm, setShowDeliveryForm] = useState(false);
  const [urlParamsApplied, setUrlParamsApplied] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [checkoutError, setCheckoutError] = useState<string | null>(null);

  // Apply URL params (from seller-generated quick link) or auto-select first variant
  // Must be above early returns to maintain consistent hook order
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    if (!data || urlParamsApplied) return;
    const variants = data.variants;
    if (variants.length === 0) return;
    const urlVariant = searchParams.get("variant");
    const urlQty = searchParams.get("qty");

    if (urlVariant && variants.some((v) => v._id === urlVariant)) {
      setSelectedVariantId(urlVariant as Id<"productVariants">);
      if (urlQty) setQuantity(Math.max(1, parseInt(urlQty) || 1));
      setShowDeliveryForm(true);
      setUrlParamsApplied(true);
    } else if (selectedVariantId === null) {
      const firstAvailable = variants.find((v) => v.stockAvailable > 0);
      if (firstAvailable) setSelectedVariantId(firstAvailable._id);
      setUrlParamsApplied(true);
    }
  }, [data, searchParams, selectedVariantId, urlParamsApplied]);

  // Loading
  if (data === undefined || isMobile === null) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-amethyst-500 border-t-transparent" />
      </div>
    );
  }

  // Not found
  if (data === null) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <h1 className="font-display text-xl font-semibold text-slate-800">
          Product not found
        </h1>
        <p className="text-slate-500">This product may have been removed.</p>
      </div>
    );
  }

  const { variants, ...product } = data;

  const selectedVariant = variants.find((v) => v._id === selectedVariantId);
  const currentPrice = selectedVariant?.priceOverride ?? product.basePrice;
  const maxQuantity = selectedVariant?.stockAvailable ?? 0;
  const totalPrice = currentPrice * quantity;
  const hasImage = product.imageUrls.length > 0;

  // Build the prefilled DM message
  const variantText = selectedVariant ? ` (${selectedVariant.variantName})` : "";
  const productUrl =
    typeof window !== "undefined"
      ? window.location.href
      : `https://swipe-social-storefront.vercel.app/shop/${storeSlug}/product/${productId}`;

  const orderMessage = `Hi! I'd like to order:\n\n${product.name}${variantText}\nQty: ${quantity}\nTotal: ${formatMVR(totalPrice)}\n\nProduct link: ${productUrl}\n\nPlease confirm availability!`;

  // Instagram DM link
  const igUsername = store?.instagramUsername;
  const igLink = igUsername
    ? `https://ig.me/m/${igUsername}`
    : null;

  // Facebook Messenger link
  const fbPageUrl = store?.facebookPageUrl;
  const fbPageId = fbPageUrl
    ? fbPageUrl.replace(/.*facebook\.com\//, "").replace(/\/$/, "")
    : null;
  const fbLink = fbPageId
    ? `https://m.me/${fbPageId}`
    : null;

  async function handleCopyMessage() {
    try {
      await navigator.clipboard.writeText(orderMessage);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      // fallback
    }
  }

  async function handleCheckout(deliveryData: DeliveryData) {
    if (!selectedVariantId) return;
    setCheckoutLoading(true);
    setCheckoutError(null);
    try {
      const res = await fetch("/api/checkout/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          productId,
          variantId: selectedVariantId,
          quantity,
          ...deliveryData,
        }),
      });
      const result = await res.json();
      if (!res.ok) {
        setCheckoutError(result.error || "Failed to create order");
        return;
      }
      router.push(`/checkout/${result.orderId}?token=${result.accessToken}`);
    } catch {
      setCheckoutError("Something went wrong. Please try again.");
    } finally {
      setCheckoutLoading(false);
    }
  }

  const hasStock = maxQuantity > 0;

  // Product info section
  const productInfo = (
    <>
      <h1 className="font-display text-xl font-bold text-slate-900 md:text-2xl">
        {product.name}
      </h1>

      <MvrAmount
        amount={currentPrice}
        className="mt-2 block text-lg text-amethyst-600 md:text-xl"
      />

      {product.description && (
        <p className="mt-3 text-sm text-slate-600 leading-relaxed">
          {product.description}
        </p>
      )}

      {/* Variant selector */}
      {variants.length > 0 && (
        <div className="mt-5">
          <label className="mb-2 block text-sm font-medium text-slate-700">
            Option
          </label>
          <VariantSelector
            variants={variants}
            selectedId={selectedVariantId}
            onSelect={(id) => {
              setSelectedVariantId(id);
              setQuantity(1);
            }}
          />
        </div>
      )}

      {/* Quantity */}
      {selectedVariant && hasStock && (
        <div className="mt-5">
          <label className="mb-2 block text-sm font-medium text-slate-700">
            Quantity
          </label>
          <QuantitySelector
            quantity={quantity}
            maxQuantity={maxQuantity}
            onChange={setQuantity}
          />
        </div>
      )}
    </>
  );

  // Reserve & Pay with Swipe section
  const reserveAndPaySection = (
    <div className="mt-6 space-y-4">
      {!showDeliveryForm ? (
        <button
          onClick={() => setShowDeliveryForm(true)}
          disabled={!hasStock || !selectedVariantId}
          className="flex w-full items-center justify-center gap-2 rounded-xl bg-amethyst-500 py-3.5 text-base font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Reserve & Pay — {formatMVR(totalPrice)}
        </button>
      ) : (
        <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-display text-base font-semibold text-slate-800">
              Delivery Details
            </h3>
            <button
              onClick={() => { setShowDeliveryForm(false); setCheckoutError(null); }}
              className="text-sm text-slate-500 hover:text-slate-700"
            >
              Cancel
            </button>
          </div>
          {checkoutError && (
            <div className="mb-4 rounded-lg bg-error/10 px-3 py-2 text-sm text-error">
              {checkoutError}
            </div>
          )}
          <DeliveryForm onSubmit={handleCheckout} isLoading={checkoutLoading} />
        </div>
      )}
    </div>
  );

  // WhatsApp link with pre-filled message
  const waNumber = store?.whatsappNumber;
  const waLink = waNumber
    ? `https://wa.me/${waNumber.replace(/\D/g, "")}?text=${encodeURIComponent(orderMessage)}`
    : null;

  // Contact seller section
  const contactSellerSection = (
    <div className="mt-4 space-y-3">
      <div className="relative flex items-center gap-3 my-2">
        <div className="flex-1 border-t border-slate-200" />
        <span className="text-xs text-slate-400 uppercase tracking-wide">or contact the seller</span>
        <div className="flex-1 border-t border-slate-200" />
      </div>

      {/* WhatsApp — pre-filled message */}
      {waLink && (
        <a
          href={waLink}
          target="_blank"
          rel="noopener noreferrer"
          className="flex w-full items-center justify-center gap-2 rounded-xl bg-[#25D366] py-3.5 text-base font-semibold text-white transition-opacity hover:opacity-90"
        >
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" />
          </svg>
          WhatsApp — Message Seller
        </a>
      )}

      {/* Instagram & Facebook — with copy disclaimer */}
      {(igLink || fbLink) && (
        <div className="space-y-2">
          <div className="flex gap-3">
            {/* Instagram DM */}
            {igLink && (
              <a
                href={igLink}
                target="_blank"
                rel="noopener noreferrer"
                className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-purple-500 via-pink-500 to-orange-400 py-3.5 text-sm font-semibold text-white transition-opacity hover:opacity-90"
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
                className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-[#0084FF] py-3.5 text-sm font-semibold text-white transition-opacity hover:opacity-90"
              >
                <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0C5.373 0 0 4.974 0 11.111c0 3.497 1.745 6.616 4.472 8.652V24l4.086-2.242c1.09.301 2.246.464 3.442.464 6.627 0 12-4.974 12-11.111S18.627 0 12 0zm1.193 14.963-3.056-3.259-5.963 3.259L10.733 8.2l3.13 3.259L19.752 8.2l-6.559 6.763z" />
                </svg>
                Facebook
              </a>
            )}
          </div>

          {/* Copy + disclaimer for IG/FB */}
          <button
            onClick={handleCopyMessage}
            disabled={!hasStock}
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-50"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v2m-6 12h8a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2h-8a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2Z" />
            </svg>
            {copied ? "Copied! Paste it in the chat" : "Copy product details"}
          </button>
          <p className="text-center text-xs text-slate-400">
            Copy the product details above and paste them in the Instagram or Facebook chat to message the seller.
          </p>
        </div>
      )}

      {/* Fallback if no social links configured */}
      {!waLink && !igLink && !fbLink && (
        <p className="text-center text-sm text-slate-400">
          Seller has not configured contact links yet.
        </p>
      )}
    </div>
  );

  // Image component
  const imageBlock = (
    <div
      className={`relative overflow-hidden bg-slate-100 ${
        isMobile ? "aspect-[4/3] w-full" : "aspect-square w-full rounded-xl"
      }`}
    >
      {hasImage ? (
        <Image
          src={product.imageUrls[0]}
          alt={product.name}
          fill
          className="object-cover"
          sizes={isMobile ? "100vw" : "50vw"}
          priority
        />
      ) : (
        <div className="flex h-full items-center justify-center text-slate-300">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
            <path strokeLinecap="round" strokeLinejoin="round" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0 0 22.5 18.75V5.25A2.25 2.25 0 0 0 20.25 3H3.75A2.25 2.25 0 0 0 1.5 5.25v13.5A2.25 2.25 0 0 0 3.75 21Z" />
          </svg>
        </div>
      )}
    </div>
  );

  if (isMobile) {
    return (
      <div className="pb-6">
        {imageBlock}
        <div className="px-4 pt-4">
          {productInfo}
          {reserveAndPaySection}
          {contactSellerSection}
        </div>
      </div>
    );
  }

  // Desktop layout
  return (
    <div className="mx-auto max-w-5xl px-6 py-8">
      <div className="grid grid-cols-2 gap-8">
        <div>{imageBlock}</div>
        <div>
          {productInfo}
          {reserveAndPaySection}
          {contactSellerSection}
        </div>
      </div>
    </div>
  );
}
