import { NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";

export async function GET() {
  try {
    const result = await convexServer.mutation(
      api.seedExchange.seedExchangeData,
      { storeSlug: "island-finds-mv" },
    );

    return NextResponse.json(result);
  } catch (err) {
    console.error("Seed exchange error:", err);
    return NextResponse.json(
      {
        error: "Failed to seed exchange data",
        details: err instanceof Error ? err.message : String(err),
      },
      { status: 500 },
    );
  }
}
