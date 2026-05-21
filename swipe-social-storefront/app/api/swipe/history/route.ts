import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getStoreIdFromSession } from "@/lib/session";
import { SESSION_COOKIE_NAME } from "@/lib/constants";
import { getTransactionHistory, getAccessTokenWithCredentials } from "@/lib/swipe-client";
import { SWIPE_API_BASE_URL } from "@/lib/constants";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";

export async function GET(request: NextRequest) {
  try {
    // Validate seller session
    const cookieStore = await cookies();
    const token = cookieStore.get(SESSION_COOKIE_NAME)?.value;

    if (!token) {
      return NextResponse.json(
        { error: "Authentication required" },
        { status: 401 },
      );
    }

    const storeId = getStoreIdFromSession(token);
    if (!storeId) {
      return NextResponse.json(
        { error: "Invalid or expired session" },
        { status: 401 },
      );
    }

    // Read pagination params from URL
    const { searchParams } = new URL(request.url);
    const limit = parseInt(searchParams.get("limit") ?? "10", 10);
    const offset = parseInt(searchParams.get("offset") ?? "0", 10);

    // Try per-store credentials first
    const store = await convexServer.query(api.stores.getById, {
      storeId: storeId as Id<"stores">,
    });

    let transactions;
    if (store?.swipeClientId && store?.swipeClientSecret) {
      const accessToken = await getAccessTokenWithCredentials(
        store.swipeClientId,
        store.swipeClientSecret,
      );
      const res = await fetch(
        `${SWIPE_API_BASE_URL}/api/v1/history?limit=${limit}&offset=${offset}`,
        { headers: { Authorization: `Bearer ${accessToken}`, Accept: "application/json" } },
      );
      if (!res.ok) throw new Error("Swipe history fetch failed");
      transactions = await res.json();
    } else {
      transactions = await getTransactionHistory(limit, offset);
    }

    return NextResponse.json(transactions);
  } catch (err) {
    console.error("Swipe getTransactionHistory error:", err);
    return NextResponse.json(
      { error: "Failed to get transaction history" },
      { status: 500 },
    );
  }
}
