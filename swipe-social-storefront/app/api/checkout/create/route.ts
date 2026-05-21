import { NextRequest, NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { createPayment, createPaymentWithCredentials } from "@/lib/swipe-client";

/** Strip HTML tags from a string for basic XSS prevention */
function stripHtml(input: string): string {
  return input.replace(/<[^>]*>/g, "");
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const {
      productId,
      variantId,
      quantity,
      customerName,
      customerPhone,
      deliveryLocation,
      deliveryAddress,
      deliveryTimePreference,
    } = body as {
      productId: string;
      variantId: string;
      quantity: number;
      customerName: string;
      customerPhone: string;
      deliveryLocation: string;
      deliveryAddress: string;
      deliveryTimePreference?: string;
    };

    // Basic validation
    if (!productId || !variantId || !quantity || !customerName || !customerPhone || !deliveryLocation || !deliveryAddress) {
      return NextResponse.json(
        { error: "Missing required fields" },
        { status: 400 },
      );
    }

    if (quantity < 1 || quantity > 99) {
      return NextResponse.json(
        { error: "Invalid quantity" },
        { status: 400 },
      );
    }

    // Sanitize free-text fields
    const safeName = stripHtml(customerName).trim();
    const safePhone = customerPhone.replace(/\D/g, "");
    const safeLocation = stripHtml(deliveryLocation).trim();
    const safeAddress = stripHtml(deliveryAddress).trim();
    const safeTimePref = deliveryTimePreference
      ? stripHtml(deliveryTimePreference).trim()
      : undefined;

    // Look up product server-side to get storeId and price
    const product = await convexServer.query(api.products.getWithVariants, {
      productId: productId as Id<"products">,
    });

    if (!product) {
      return NextResponse.json(
        { error: "Product not found" },
        { status: 404 },
      );
    }

    // Create order in Convex
    const orderResult = await convexServer.mutation(api.orders.createOrder, {
      storeId: product.storeId,
      productId: productId as Id<"products">,
      variantId: variantId as Id<"productVariants">,
      quantity,
      customerName: safeName,
      customerPhone: safePhone,
      deliveryLocation: safeLocation,
      deliveryAddress: safeAddress,
      deliveryTimePreference: safeTimePref,
    });

    if (!orderResult.success) {
      const statusCode = orderResult.reason === "insufficient_stock" ? 409 : 400;
      return NextResponse.json(
        { error: orderResult.reason },
        { status: statusCode },
      );
    }

    const { orderId, accessToken } = orderResult;

    // Determine amount from the variant
    const variant = product.variants.find(
      (v) => v._id === (variantId as Id<"productVariants">),
    );
    const unitPrice = variant?.priceOverride ?? product.basePrice;
    const totalAmount = unitPrice * quantity;

    // Fetch store to check for per-store Swipe credentials
    const store = await convexServer.query(api.stores.getById, {
      storeId: product.storeId,
    });

    // Create Swipe payment — always use createPayment which handles demo mode
    // internally. Only use per-store credentials when NOT in demo mode.
    const swipePayment = await createPayment(
      totalAmount,
      product.currency,
      `Order for ${product.name}`,
    );

    // Update order with Swipe data
    await convexServer.mutation(api.orders.setSwipeData, {
      orderId: orderId as Id<"orders">,
      swipePaymentId: swipePayment.id,
      swipeReference: swipePayment.reference,
      swipeShortCode: swipePayment.short_code,
      swipeQrData: swipePayment.qr_data,
      swipePaymentUrl: swipePayment.payment_url,
    });

    return NextResponse.json({
      orderId,
      accessToken,
      paymentId: swipePayment.id,
      shortCode: swipePayment.short_code,
      qrData: swipePayment.qr_data,
      paymentUrl: swipePayment.payment_url,
    });
  } catch (err) {
    console.error("Checkout create error:", err);
    return NextResponse.json(
      { error: "Failed to create order" },
      { status: 500 },
    );
  }
}
