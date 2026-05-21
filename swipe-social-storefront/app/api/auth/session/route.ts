import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getStoreIdFromSession } from "@/lib/session";
import { SESSION_COOKIE_NAME, SESSION_TTL_MS } from "@/lib/constants";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import type { Id } from "@/convex/_generated/dataModel";

/**
 * GET /api/auth/session
 * Bootstraps a Convex session from the existing httpOnly seller_session cookie.
 * Sets the client-readable seller_session_token cookie.
 * Called by seller layout on first load if seller_session_token is missing.
 */
export async function GET(request: NextRequest) {
  try {
    const cookieStore = await cookies();
    const token = cookieStore.get(SESSION_COOKIE_NAME)?.value;

    if (!token) {
      return NextResponse.json({ error: "Not authenticated" }, { status: 401 });
    }

    const storeId = getStoreIdFromSession(token);
    if (!storeId) {
      return NextResponse.json({ error: "Invalid session" }, { status: 401 });
    }

    // Create Convex session if it doesn't exist
    try {
      await convexServer.mutation(api.sessions.create, {
        token,
        storeId: storeId as Id<"stores">,
      });
    } catch {
      // Session may already exist, ignore
    }

    // Set client-readable cookie
    const response = NextResponse.json({ ok: true });
    response.cookies.set("seller_session_token", token, {
      httpOnly: false,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: SESSION_TTL_MS / 1000,
    });

    return response;
  } catch {
    return NextResponse.json({ error: "Session bootstrap failed" }, { status: 500 });
  }
}
