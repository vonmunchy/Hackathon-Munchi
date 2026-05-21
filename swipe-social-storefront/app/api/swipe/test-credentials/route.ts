import { NextRequest, NextResponse } from "next/server";
import { SWIPE_API_BASE_URL } from "@/lib/constants";

export async function POST(request: NextRequest) {
  try {
    const { clientId, clientSecret } = await request.json();

    if (!clientId || !clientSecret) {
      return NextResponse.json(
        { valid: false, error: "Client ID and Secret are required" },
        { status: 400 },
      );
    }

    const res = await fetch(`${SWIPE_API_BASE_URL}/oauth2/token`, {
      method: "POST",
      headers: {
        Authorization: `Basic ${btoa(clientId + ":" + clientSecret)}`,
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: "grant_type=client_credentials",
    });

    if (!res.ok) {
      return NextResponse.json({ valid: false, error: "Invalid credentials" });
    }

    const data = await res.json();
    return NextResponse.json({
      valid: true,
      message: `Connected! Token expires in ${data.expires_in}s`,
    });
  } catch {
    return NextResponse.json(
      { valid: false, error: "Could not connect to Swipe API" },
      { status: 500 },
    );
  }
}
