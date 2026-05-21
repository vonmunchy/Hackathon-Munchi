/**
 * Singleton Swipe API client.
 * All Swipe API calls go through this module — the browser never calls Swipe directly.
 * Bearer tokens are cached server-side and never exposed to the client.
 */

import {
  SWIPE_API_BASE_URL,
  SWIPE_DEMO_MODE,
  SWIPE_CLIENT_ID,
  SWIPE_CLIENT_SECRET,
} from "./constants";
import {
  createDemoPayment,
  getDemoPaymentStatus,
  simulateDemoCompletion,
  simulateDemoExpiry,
  simulateDemoCancel,
  getDemoBalance,
  getDemoHistory,
} from "./swipe-demo";

// ---------- Demo Logging ----------
const RESET = "\x1b[0m";
const GREEN = "\x1b[32m";
const YELLOW = "\x1b[33m";
const RED = "\x1b[31m";
const CYAN = "\x1b[36m";
const BOLD = "\x1b[1m";
const DIM = "\x1b[2m";

function timestamp(): string {
  return new Date().toLocaleTimeString("en-US", { hour12: false });
}

function swipeLog(
  level: "success" | "pending" | "error" | "info",
  operation: string,
  details: string,
): void {
  const icons = { success: "✓", pending: "⏳", error: "✗", info: "ℹ" };
  const colors = { success: GREEN, pending: YELLOW, error: RED, info: CYAN };
  const color = colors[level];
  const icon = icons[level];
  console.log(
    `${DIM}[SWIPE]${RESET} ${DIM}${timestamp()}${RESET} ${color}${BOLD}${icon}${RESET} ${color}${operation}${RESET}${details ? ` — ${details}` : ""}`,
  );
}

// ---------- Token cache (per-credential) ----------
const tokenCache = new Map<string, { token: string; expiresAt: number }>();

function cacheKey(clientId: string, clientSecret: string): string {
  return `${clientId}:${clientSecret}`;
}

/**
 * Returns a valid Bearer access token for the given credentials,
 * refreshing from the OAuth2 client-credentials endpoint when
 * the cached token is missing or within 60 s of expiry.
 */
export async function getAccessTokenWithCredentials(
  clientId: string,
  clientSecret: string,
): Promise<string> {
  const key = cacheKey(clientId, clientSecret);
  const cached = tokenCache.get(key);

  if (cached && Date.now() < cached.expiresAt - 60_000) {
    return cached.token;
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
    const text = await res.text();
    throw new Error(`Swipe OAuth2 token exchange failed (${res.status}): ${text}`);
  }

  const data: { access_token: string; token_type: string; expires_in: number } =
    await res.json();

  tokenCache.set(key, {
    token: data.access_token,
    expiresAt: Date.now() + data.expires_in * 1000,
  });

  return data.access_token;
}

/** Backward-compatible: uses env var credentials */
export async function getAccessToken(): Promise<string> {
  return getAccessTokenWithCredentials(SWIPE_CLIENT_ID, SWIPE_CLIENT_SECRET);
}

// ---------- Payments ----------

export interface SwipePaymentResponse {
  id: string;
  amount: number;
  currency: string;
  status: string;
  reference?: string;
  short_code?: string;
  qr_data?: string;
  payment_url?: string;
  created_at?: string;
}

export async function createPaymentWithCredentials(
  amount: number,
  currency: string,
  description: string,
  clientId: string,
  clientSecret: string,
): Promise<SwipePaymentResponse> {
  const token = await getAccessTokenWithCredentials(clientId, clientSecret);
  const res = await fetch(`${SWIPE_API_BASE_URL}/api/v1/payments`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ amount, currency, type: "QR", description }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Swipe createPayment failed (${res.status}): ${text}`);
  }

  return res.json();
}

export async function createPayment(
  amount: number,
  currency: string,
  description: string,
): Promise<SwipePaymentResponse> {
  swipeLog("pending", "Payment CREATE", `${currency} ${amount.toFixed(2)} — "${description}"`);
  try {
    let result: SwipePaymentResponse;
    if (SWIPE_DEMO_MODE) {
      result = createDemoPayment(amount, description);
    } else {
      result = await createPaymentWithCredentials(amount, currency, description, SWIPE_CLIENT_ID, SWIPE_CLIENT_SECRET);
    }
    swipeLog("success", "Payment CREATED", `${currency} ${result.amount?.toFixed(2) ?? amount.toFixed(2)} — ID: ${result.id}${result.short_code ? ` — Code: ${result.short_code}` : ""}`);
    return result;
  } catch (err) {
    swipeLog("error", "Payment CREATE FAILED", err instanceof Error ? err.message : String(err));
    throw err;
  }
}

export async function getPaymentStatus(
  paymentId: string,
): Promise<SwipePaymentResponse> {
  swipeLog("pending", "Status CHECK", `Payment: ${paymentId}`);
  try {
    let result: SwipePaymentResponse;
    if (SWIPE_DEMO_MODE) {
      const payment = getDemoPaymentStatus(paymentId);
      if (!payment) throw new Error(`Demo payment not found: ${paymentId}`);
      result = payment;
    } else {
      const token = await getAccessToken();
      const res = await fetch(
        `${SWIPE_API_BASE_URL}/api/v1/payments/${paymentId}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/json",
          },
        },
      );

      if (!res.ok) {
        const text = await res.text();
        throw new Error(`Swipe getPaymentStatus failed (${res.status}): ${text}`);
      }

      result = await res.json();
    }
    swipeLog("success", "Status CHECK", `Payment: ${paymentId} — Status: ${result.status?.toUpperCase()}`);
    return result;
  } catch (err) {
    swipeLog("error", "Status CHECK FAILED", `Payment: ${paymentId} — ${err instanceof Error ? err.message : String(err)}`);
    throw err;
  }
}

// ---------- Simulate (mock pay-page) ----------

export async function simulatePayment(
  shortCode: string,
  action: string,
): Promise<void> {
  swipeLog("pending", `Simulate ${action.toUpperCase()}`, `Code: ${shortCode}`);
  try {
    if (SWIPE_DEMO_MODE) {
      if (action === "complete") simulateDemoCompletion(shortCode);
      else if (action === "expire") simulateDemoExpiry(shortCode);
      else if (action === "cancel") simulateDemoCancel(shortCode);
    } else {
      const res = await fetch(
        `${SWIPE_API_BASE_URL}/pay/${shortCode}/${action}`,
        {
          method: "POST",
          redirect: "manual",
        },
      );

      if (res.status !== 303 && !res.ok) {
        const text = await res.text().catch(() => "");
        throw new Error(
          `Swipe simulatePayment failed (${res.status}): ${text}`,
        );
      }
    }
    swipeLog("success", `Simulate ${action.toUpperCase()}`, `Code: ${shortCode} — Done`);
  } catch (err) {
    swipeLog("error", `Simulate ${action.toUpperCase()} FAILED`, `Code: ${shortCode} — ${err instanceof Error ? err.message : String(err)}`);
    throw err;
  }
}

// ---------- Wallet / Balance ----------

export async function getWalletBalance(): Promise<
  Array<{ available_balance: number; pending_balance: number; currency: string }>
> {
  swipeLog("pending", "Wallet BALANCE", "Fetching...");
  try {
    let result: Array<{ available_balance: number; pending_balance: number; currency: string }>;
    if (SWIPE_DEMO_MODE) {
      result = getDemoBalance();
    } else {
      const token = await getAccessToken();
      const res = await fetch(`${SWIPE_API_BASE_URL}/api/v1/balance`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
          Accept: "application/json",
        },
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(`Swipe getWalletBalance failed (${res.status}): ${text}`);
      }

      result = await res.json();
    }
    const balanceStr = result.map(b => `${b.currency} ${b.available_balance.toFixed(2)}`).join(", ");
    swipeLog("success", "Wallet BALANCE", balanceStr);
    return result;
  } catch (err) {
    swipeLog("error", "Wallet BALANCE FAILED", err instanceof Error ? err.message : String(err));
    throw err;
  }
}

// ---------- Transaction History ----------

export async function getTransactionHistory(
  limit: number,
  offset: number,
): Promise<{ transactions: unknown[]; total: number }> {
  swipeLog("pending", "Transaction HISTORY", `limit=${limit} offset=${offset}`);
  try {
    let result: { transactions: unknown[]; total: number };
    if (SWIPE_DEMO_MODE) {
      result = getDemoHistory();
    } else {
      const token = await getAccessToken();
      const res = await fetch(
        `${SWIPE_API_BASE_URL}/api/v1/history?limit=${limit}&offset=${offset}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/json",
          },
        },
      );

      if (!res.ok) {
        const text = await res.text();
        throw new Error(
          `Swipe getTransactionHistory failed (${res.status}): ${text}`,
        );
      }

      result = await res.json();
    }
    swipeLog("success", "Transaction HISTORY", `${result.total} total, ${result.transactions.length} returned`);
    return result;
  } catch (err) {
    swipeLog("error", "Transaction HISTORY FAILED", err instanceof Error ? err.message : String(err));
    throw err;
  }
}
