import { TIMEZONE_OFFSET_HOURS } from "./constants";

/** Format a number as "MVR X,XXX.XX" */
export function formatMVR(amount: number): string {
  return (
    "MVR " +
    amount
      .toFixed(2)
      .replace(/\B(?=(\d{3})+(?!\d))/g, ",")
  );
}

/** Convert a UTC timestamp to Maldives time (UTC+5) and format */
export function formatMaldivesTime(timestamp: number): string {
  const date = new Date(timestamp + TIMEZONE_OFFSET_HOURS * 60 * 60 * 1000);
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  const hours = String(date.getUTCHours()).padStart(2, "0");
  const minutes = String(date.getUTCMinutes()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

/** Format a short order ID from a Convex ID string */
export function shortOrderId(id: string): string {
  // Convex IDs are like "k17abc123..." — take last 8 chars
  return id.slice(-8);
}
