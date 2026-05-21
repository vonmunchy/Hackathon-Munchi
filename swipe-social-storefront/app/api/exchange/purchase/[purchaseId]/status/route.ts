import { NextRequest, NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { getPaymentStatus } from "@/lib/swipe-client";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ purchaseId: string }> },
) {
  try {
    const { purchaseId } = await params;

    if (!purchaseId) {
      return NextResponse.json(
        { error: "purchaseId is required" },
        { status: 400 },
      );
    }

    // Get paymentId from query params
    const { searchParams } = new URL(request.url);
    const paymentId = searchParams.get("paymentId");

    if (!paymentId) {
      return NextResponse.json(
        { error: "paymentId query parameter is required" },
        { status: 400 },
      );
    }

    // Check Swipe payment status (handles demo mode internally)
    const paymentStatus = await getPaymentStatus(paymentId);

    if (paymentStatus.status === "COMPLETED") {
      // Payment completed — confirm purchase in Convex
      try {
        const result = await convexServer.mutation(api.exchange.confirmPurchase, {
          reservationId: purchaseId as Id<"exchangeReservations">,
          swipePaymentId: paymentId,
        });

        return NextResponse.json({
          status: "completed",
          txHash: result.txHash,
          usdtAmount: result.usdtAmount,
          mvrAmount: result.mvrAmount,
        });
      } catch (err: unknown) {
        // If reservation already completed, still return completed status
        const message = err instanceof Error ? err.message : "";
        if (message.includes("not active")) {
          return NextResponse.json({
            status: "completed",
          });
        }
        throw err;
      }
    }

    if (
      paymentStatus.status === "EXPIRED" ||
      paymentStatus.status === "CANCELLED"
    ) {
      return NextResponse.json({ status: "failed" });
    }

    // Still pending
    return NextResponse.json({ status: "pending" });
  } catch (err) {
    console.error("Exchange purchase status error:", err);
    return NextResponse.json(
      { error: "Failed to check purchase status" },
      { status: 500 },
    );
  }
}
