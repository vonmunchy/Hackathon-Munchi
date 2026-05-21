import { NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";

const ADMIN_PASSPHRASE = process.env.ADMIN_PASSPHRASE || "swipe2026";

function checkAuth(request: Request): boolean {
  const auth = request.headers.get("Authorization");
  return auth === ADMIN_PASSPHRASE;
}

export async function GET(request: Request) {
  if (!checkAuth(request)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const stores = await convexServer.query(api.stores.listAll, {});
    return NextResponse.json(stores);
  } catch (err) {
    console.error("Admin listAll error:", err);
    return NextResponse.json(
      { error: "Failed to fetch stores" },
      { status: 500 },
    );
  }
}
