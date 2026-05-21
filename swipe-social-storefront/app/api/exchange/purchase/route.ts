import { NextRequest, NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";
import { createPayment } from "@/lib/swipe-client";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { listingId, usdtAmount, buyerWallet } = body as {
      listingId: string;
      usdtAmount: number;
      buyerWallet: string;
    };

    // Validate required fields
    if (!listingId || !usdtAmount || !buyerWallet) {
      return NextResponse.json(
        { error: "Missing required fields: listingId, usdtAmount, buyerWallet" },
        { status: 400 },
      );
    }

    // Validate buyerWallet format: starts with "T", exactly 34 chars, alphanumeric
    if (
      !buyerWallet.startsWith("T") ||
      buyerWallet.length !== 34 ||
      !/^[A-Za-z0-9]+$/.test(buyerWallet)
    ) {
      return NextResponse.json(
        { error: "Invalid TRC20 wallet address: must start with T, be 34 characters, and alphanumeric" },
        { status: 400 },
      );
    }

    if (usdtAmount <= 0) {
      return NextResponse.json(
        { error: "usdtAmount must be greater than 0" },
        { status: 400 },
      );
    }

    // Fetch listing details to get rate
    const listing = await convexServer.query(api.exchange.getListingDetails, {
      listingId: listingId as Id<"exchangeListings">,
    });

    if (!listing) {
      return NextResponse.json(
        { error: "Listing not found" },
        { status: 404 },
      );
    }

    if (listing.status !== "active") {
      return NextResponse.json(
        { error: "Listing is not active" },
        { status: 400 },
      );
    }

    if (usdtAmount > listing.availableBalance) {
      return NextResponse.json(
        { error: "Insufficient available balance on listing" },
        { status: 409 },
      );
    }

    // Create reservation in Convex
    let reservationResult;
    try {
      reservationResult = await convexServer.mutation(api.exchange.createReservation, {
        listingId: listingId as Id<"exchangeListings">,
        amount: usdtAmount,
        buyerWallet,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to create reservation";
      if (message.includes("Insufficient")) {
        return NextResponse.json({ error: message }, { status: 409 });
      }
      if (message.includes("not found")) {
        return NextResponse.json({ error: message }, { status: 404 });
      }
      return NextResponse.json({ error: message }, { status: 400 });
    }

    // Calculate MVR amount
    const mvrAmount = usdtAmount * listing.rate;

    // Create Swipe payment (createPayment already handles demo mode internally)
    const swipePayment = await createPayment(
      mvrAmount,
      "MVR",
      `USDT Purchase - ${usdtAmount} USDT`,
    );

    return NextResponse.json({
      reservationId: reservationResult.reservationId,
      paymentId: swipePayment.id,
      paymentUrl: swipePayment.payment_url,
      shortCode: swipePayment.short_code,
      qrData: swipePayment.qr_data,
      mvrAmount,
      usdtAmount,
    });
  } catch (err) {
    console.error("Exchange purchase error:", err);
    return NextResponse.json(
      { error: "Failed to initiate purchase" },
      { status: 500 },
    );
  }
}
