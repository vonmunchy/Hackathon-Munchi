"use client";

import { useState, useEffect } from "react";

/**
 * Returns the seller session token from the client-readable cookie.
 * Used to pass to Convex queries/mutations for multi-tenant auth.
 */
export function useSessionToken(): string | null {
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    const match = document.cookie
      .split(";")
      .map((c) => c.trim())
      .find((c) => c.startsWith("seller_session_token="));
    if (match) {
      setToken(decodeURIComponent(match.split("=")[1]));
    }
  }, []);

  return token;
}
