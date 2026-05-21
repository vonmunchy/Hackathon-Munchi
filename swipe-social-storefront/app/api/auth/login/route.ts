import { NextRequest, NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import { createSession } from "@/lib/session";
import { SESSION_COOKIE_NAME, SESSION_TTL_MS } from "@/lib/constants";

export async function POST(request: NextRequest) {
  try {
    const { slug, pin } = await request.json();

    if (!slug || !pin) {
      return NextResponse.json(
        { error: "Store ID and PIN are required" },
        { status: 400 },
      );
    }

    const result = await convexServer.query(api.stores.verifyPin, { slug, pin });

    if (!result.found) {
      return NextResponse.json(
        { error: "Store not found", field: "slug" },
        { status: 401 },
      );
    }

    if (!result.pinMatch) {
      return NextResponse.json(
        { error: "Incorrect PIN", field: "pin" },
        { status: 401 },
      );
    }

    // Create session and set cookie
    const token = createSession(result.storeId as string);

    // Get store details for the response
    const store = await convexServer.query(api.stores.getBySlug, { slug });

    const response = NextResponse.json({
      storeSlug: slug,
      storeName: store?.name ?? slug,
    });

    response.cookies.set(SESSION_COOKIE_NAME, token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    // Non-httpOnly flag cookie for client-side auth guard (layout redirect)
    response.cookies.set("seller_logged_in", "1", {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    return response;
  } catch (err) {
    console.error("Login error:", err);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}
