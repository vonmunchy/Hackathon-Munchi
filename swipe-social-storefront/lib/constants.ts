export const SWIPE_API_BASE_URL =
  process.env.SWIPE_API_BASE_URL || "http://127.0.0.1:8080";

export const SWIPE_DEMO_MODE =
  process.env.SWIPE_DEMO_MODE === "true";

export const SWIPE_CLIENT_ID = process.env.SWIPE_CLIENT_ID || "";
export const SWIPE_CLIENT_SECRET = process.env.SWIPE_CLIENT_SECRET || "";

export const CURRENCY = "MVR";
export const TIMEZONE_OFFSET_HOURS = 5; // Maldives UTC+5
export const TIMEZONE_LABEL = "MVT"; // Maldives Time

export const STORE_DEMO_SLUG = "island-finds-mv";
export const STORE_DEMO_PIN = "1234";

export const SESSION_COOKIE_NAME = "seller_session";
export const SESSION_TTL_MS = 24 * 60 * 60 * 1000; // 24 hours
