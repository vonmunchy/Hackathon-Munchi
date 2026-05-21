import { NextRequest, NextResponse } from "next/server";
import { simulatePayment } from "@/lib/swipe-client";

const VALID_ACTIONS = ["complete", "expire", "cancel"];

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { shortCode, action } = body as {
      shortCode: string;
      action: string;
    };

    if (!shortCode || !action) {
      return NextResponse.json(
        { error: "shortCode and action are required" },
        { status: 400 },
      );
    }

    if (!VALID_ACTIONS.includes(action)) {
      return NextResponse.json(
        { error: `action must be one of: ${VALID_ACTIONS.join(", ")}` },
        { status: 400 },
      );
    }

    await simulatePayment(shortCode, action);

    return NextResponse.json({ ok: true });
  } catch (err) {
    console.error("Swipe simulatePayment error:", err);
    return NextResponse.json(
      { error: "Failed to simulate payment" },
      { status: 500 },
    );
  }
}
