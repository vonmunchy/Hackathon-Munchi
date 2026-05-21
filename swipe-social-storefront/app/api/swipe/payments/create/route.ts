import { NextRequest, NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import { createPayment } from "@/lib/swipe-client";
import type { Id } from "@/convex/_generated/dataModel";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { orderId } = body as { orderId: string };

    if (!orderId) {
      return NextResponse.json(
        { error: "orderId is required" },
        { status: 400 },
      );
    }

    // Look up order in Convex to get amount and product details
    const result = await convexServer.query(api.orders.getOrderWithDetails, {
      orderId: orderId as Id<"orders">,
    });

    if (!result) {
      return NextResponse.json(
        { error: "Order not found" },
        { status: 404 },
      );
    }

    const { order, product } = result;
    const productName = product?.name ?? "Order";

    // Create Swipe payment
    const payment = await createPayment(
      order.totalAmount,
      order.currency,
      `${productName} - Order`,
    );

    // Store Swipe data back on the Convex order
    await convexServer.mutation(api.orders.setSwipeData, {
      orderId: orderId as Id<"orders">,
      swipePaymentId: payment.id,
      swipeReference: payment.reference,
      swipeShortCode: payment.short_code,
      swipeQrData: payment.qr_data,
      swipePaymentUrl: payment.payment_url,
    });

    return NextResponse.json({
      paymentId: payment.id,
      shortCode: payment.short_code,
      qrData: payment.qr_data,
      paymentUrl: payment.payment_url,
    });
  } catch (err) {
    console.error("Swipe createPayment error:", err);
    return NextResponse.json(
      { error: "Failed to create payment" },
      { status: 500 },
    );
  }
}
