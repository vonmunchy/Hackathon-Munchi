/**
 * Demo-mode mock for the Swipe payment API.
 * Used when SWIPE_DEMO_MODE=true so the app works without the real mock server.
 */

export interface DemoPayment {
  id: string;
  amount: number;
  currency: string;
  status: "PENDING" | "COMPLETED" | "EXPIRED" | "CANCELLED";
  reference: string;
  short_code: string;
  qr_data: string;
  payment_url: string;
  description: string;
  created_at: string;
}

// In-memory store for demo payment state
const demoPayments = new Map<string, DemoPayment>();

// Minimal valid 1x1 white PNG as base64
const PLACEHOLDER_QR_BASE64 =
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==";

function randomShortCode(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  let code = "";
  for (let i = 0; i < 6; i++) {
    code += chars[Math.floor(Math.random() * chars.length)];
  }
  return code;
}

export function createDemoPayment(
  amount: number,
  description: string,
): DemoPayment {
  const id = `pay_demo_${crypto.randomUUID()}`;
  const shortCode = `DEMO-${randomShortCode()}`;
  const now = new Date().toISOString();

  const payment: DemoPayment = {
    id,
    amount,
    currency: "MVR",
    status: "PENDING",
    reference: shortCode,
    short_code: shortCode,
    qr_data: PLACEHOLDER_QR_BASE64,
    payment_url: `#demo-pay`,
    description,
    created_at: now,
  };

  demoPayments.set(id, payment);
  // Also index by short_code for simulate lookups
  demoPayments.set(shortCode, payment);

  return payment;
}

export function getDemoPaymentStatus(paymentId: string): DemoPayment | null {
  return demoPayments.get(paymentId) ?? null;
}

export function simulateDemoCompletion(paymentId: string): boolean {
  const payment = demoPayments.get(paymentId);
  if (!payment) return false;
  payment.status = "COMPLETED";
  return true;
}

export function simulateDemoExpiry(paymentId: string): boolean {
  const payment = demoPayments.get(paymentId);
  if (!payment) return false;
  payment.status = "EXPIRED";
  return true;
}

export function simulateDemoCancel(paymentId: string): boolean {
  const payment = demoPayments.get(paymentId);
  if (!payment) return false;
  payment.status = "CANCELLED";
  return true;
}

export async function* streamDemoPayment(
  paymentId: string,
): AsyncGenerator<string, void, unknown> {
  const payment = demoPayments.get(paymentId);
  if (!payment) return;

  // Yield initial PENDING state
  yield `data: ${JSON.stringify({ id: payment.id, status: payment.status, timestamp: new Date().toISOString() })}\n\n`;

  const startTime = Date.now();
  const AUTO_COMPLETE_MS = 3000;
  const POLL_INTERVAL_MS = 500;

  while (true) {
    await new Promise((resolve) => setTimeout(resolve, POLL_INTERVAL_MS));

    const current = demoPayments.get(paymentId);
    if (!current) return;

    // Auto-complete after 3 seconds if still pending and no explicit action
    if (
      current.status === "PENDING" &&
      Date.now() - startTime >= AUTO_COMPLETE_MS
    ) {
      current.status = "COMPLETED";
    }

    yield `data: ${JSON.stringify({ id: current.id, status: current.status, timestamp: new Date().toISOString() })}\n\n`;

    // Terminal states end the stream
    if (
      current.status === "COMPLETED" ||
      current.status === "EXPIRED" ||
      current.status === "CANCELLED"
    ) {
      return;
    }
  }
}

export function getDemoBalance() {
  return [
    {
      available_balance: 15420.5,
      pending_balance: 650.0,
      currency: "MVR",
    },
  ];
}

export function getDemoHistory() {
  const now = new Date();
  return {
    transactions: [
      {
        id: "demo-txn-001",
        reference: "TXN-DEMO-A1B2C3",
        amount: 250.0,
        currency: "MVR",
        type: "P2M",
        status: "COMPLETED",
        description: "Handmade Bracelet - Order",
        gross_amount: 250.0,
        fee_amount: 5.0,
        net_amount: 245.0,
        created_at: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
      },
      {
        id: "demo-txn-002",
        reference: "TXN-DEMO-D4E5F6",
        amount: 180.0,
        currency: "MVR",
        type: "P2M",
        status: "COMPLETED",
        description: "Shell Necklace - Order",
        gross_amount: 180.0,
        fee_amount: 3.6,
        net_amount: 176.4,
        created_at: new Date(
          now.getTime() - 24 * 60 * 60 * 1000,
        ).toISOString(),
      },
      {
        id: "demo-txn-003",
        reference: "TXN-DEMO-G7H8I9",
        amount: 420.5,
        currency: "MVR",
        type: "P2M",
        status: "COMPLETED",
        description: "Coral Earrings Set - Order",
        gross_amount: 420.5,
        fee_amount: 8.41,
        net_amount: 412.09,
        created_at: new Date(
          now.getTime() - 3 * 24 * 60 * 60 * 1000,
        ).toISOString(),
      },
    ],
    total: 3,
  };
}
