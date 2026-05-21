import { NextRequest, NextResponse } from "next/server";
import { getPaymentStatus } from "@/lib/swipe-client";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ paymentId: string }> },
) {
  try {
    const { paymentId } = await params;

    if (!paymentId) {
      return NextResponse.json(
        { error: "paymentId is required" },
        { status: 400 },
      );
    }

    const status = await getPaymentStatus(paymentId);

    return NextResponse.json(status);
  } catch (err) {
    console.error("Swipe getPaymentStatus error:", err);
    return NextResponse.json(
      { error: "Failed to get payment status" },
      { status: 500 },
    );
  }
}
