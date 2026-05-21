import { NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import { Id } from "@/convex/_generated/dataModel";

const ADMIN_PASSPHRASE = process.env.ADMIN_PASSPHRASE || "swipe2026";

function checkAuth(request: Request): boolean {
  const auth = request.headers.get("Authorization");
  return auth === ADMIN_PASSPHRASE;
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ storeId: string }> },
) {
  if (!checkAuth(request)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const { storeId } = await params;

  try {
    const result = await convexServer.mutation(api.stores.resetStore, {
      storeId: storeId as Id<"stores">,
    });
    return NextResponse.json(result);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    if (message.includes("not found")) {
      return NextResponse.json({ error: "Store not found" }, { status: 404 });
    }
    console.error("Admin resetStore error:", err);
    return NextResponse.json(
      { error: "Failed to reset store" },
      { status: 500 },
    );
  }
}
