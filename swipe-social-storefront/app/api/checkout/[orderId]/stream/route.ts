import { NextRequest } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import { getAccessToken } from "@/lib/swipe-client";
import { SWIPE_API_BASE_URL, SWIPE_DEMO_MODE } from "@/lib/constants";
import { streamDemoPayment } from "@/lib/swipe-demo";
import type { Id } from "@/convex/_generated/dataModel";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ orderId: string }> },
) {
  const { orderId } = await params;
  const token = request.nextUrl.searchParams.get("token");

  if (!orderId || !token) {
    return new Response("Missing orderId or token", { status: 400 });
  }

  // Validate access token against order
  const order = await convexServer.query(api.orders.getByIdWithToken, {
    orderId: orderId as Id<"orders">,
    accessToken: token,
  });

  if (!order) {
    return new Response("Forbidden", { status: 403 });
  }

  if (!order.swipePaymentId) {
    return new Response("Payment not yet created", { status: 404 });
  }

  const headers = {
    "Content-Type": "text/event-stream",
    "Cache-Control": "no-cache",
    Connection: "keep-alive",
  };

  // Demo mode: simulate SSE
  if (SWIPE_DEMO_MODE) {
    const stream = new ReadableStream({
      async start(controller) {
        try {
          for await (const chunk of streamDemoPayment(order.swipePaymentId!)) {
            controller.enqueue(new TextEncoder().encode(chunk));

            // Check for terminal status and confirm payment
            const data = JSON.parse(chunk.replace("data: ", "").trim());
            if (data.status === "COMPLETED") {
              await convexServer.mutation(api.orders.confirmPaymentPublic, {
                orderId: orderId as Id<"orders">,
              });
            }
          }
        } catch {
          // Stream closed
        } finally {
          controller.close();
        }
      },
    });

    return new Response(stream, { headers });
  }

  // Real mode: proxy Swipe SSE
  const bearerToken = await getAccessToken();
  const upstreamUrl = `${SWIPE_API_BASE_URL}/api/v1/payments/${order.swipePaymentId}/stream`;

  const upstream = await fetch(upstreamUrl, {
    headers: {
      Authorization: `Bearer ${bearerToken}`,
      Accept: "text/event-stream",
    },
    signal: request.signal,
  });

  if (!upstream.ok || !upstream.body) {
    return new Response("Failed to connect to payment stream", { status: 502 });
  }

  const reader = upstream.body.getReader();
  const decoder = new TextDecoder();

  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder();
      let buffer = "";

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });

          // Process complete SSE events (double newline separated)
          const events = buffer.split("\n\n");
          buffer = events.pop() || "";

          for (const event of events) {
            if (!event.trim()) continue;

            // Re-emit the event to the browser
            controller.enqueue(encoder.encode(event + "\n\n"));

            // Check for terminal status in data lines
            const dataLine = event
              .split("\n")
              .find((l) => l.startsWith("data:"));
            if (dataLine) {
              try {
                const data = JSON.parse(dataLine.slice(5).trim());
                if (data.status === "COMPLETED") {
                  await convexServer.mutation(
                    api.orders.confirmPaymentPublic,
                    { orderId: orderId as Id<"orders"> },
                  );
                }
              } catch {
                // Parse error — skip
              }
            }
          }
        }
      } catch {
        // Stream closed by client or upstream
      } finally {
        reader.releaseLock();
        controller.close();
      }
    },
    cancel() {
      reader.cancel();
    },
  });

  return new Response(stream, { headers });
}
