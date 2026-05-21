"use client";

import { useState, useEffect, useRef, useCallback } from "react";

type PaymentStatus =
  | "PENDING"
  | "COMPLETED"
  | "EXPIRED"
  | "CANCELLED"
  | "UNKNOWN";

interface UsePaymentStatusOptions {
  orderId: string;
  accessToken: string;
  swipePaymentId: string | null;
  enabled?: boolean;
}

interface UsePaymentStatusResult {
  status: PaymentStatus;
  isPolling: boolean;
  error: string | null;
}

const MAX_SSE_RETRIES = 3;
const POLL_INTERVAL_MS = 5000;
const MAX_POLL_DURATION_MS = 5 * 60 * 1000; // 5 minutes

export function usePaymentStatus({
  orderId,
  accessToken,
  swipePaymentId,
  enabled = true,
}: UsePaymentStatusOptions): UsePaymentStatusResult {
  const [status, setStatus] = useState<PaymentStatus>("PENDING");
  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sseRetriesRef = useRef(0);
  const eventSourceRef = useRef<EventSource | null>(null);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const pollStartRef = useRef<number>(0);

  const isTerminal = useCallback(
    (s: PaymentStatus) =>
      s === "COMPLETED" || s === "EXPIRED" || s === "CANCELLED",
    [],
  );

  const cleanup = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current);
      pollIntervalRef.current = null;
    }
  }, []);

  const startPolling = useCallback(() => {
    if (pollIntervalRef.current) return;
    setIsPolling(true);
    pollStartRef.current = Date.now();

    pollIntervalRef.current = setInterval(async () => {
      if (Date.now() - pollStartRef.current > MAX_POLL_DURATION_MS) {
        cleanup();
        setError("Payment status check timed out");
        return;
      }

      try {
        const res = await fetch(
          `/api/swipe/payments/${swipePaymentId}/status`,
        );
        if (res.ok) {
          const data = await res.json();
          const newStatus = data.status as PaymentStatus;
          setStatus(newStatus);
          if (isTerminal(newStatus)) {
            cleanup();
          }
        }
      } catch {
        // Continue polling on error
      }
    }, POLL_INTERVAL_MS);
  }, [swipePaymentId, cleanup, isTerminal]);

  const startSSE = useCallback(() => {
    if (!orderId || !accessToken) return;

    const url = `/api/checkout/${orderId}/stream?token=${accessToken}`;
    const es = new EventSource(url);
    eventSourceRef.current = es;

    // Handle unnamed events (onmessage)
    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const newStatus = data.status as PaymentStatus;
        setStatus(newStatus);
        if (isTerminal(newStatus)) {
          cleanup();
        }
      } catch {
        // Parse error
      }
    };

    // Also handle named "status" events (Swipe mock may emit these)
    es.addEventListener("status", (event) => {
      try {
        const data = JSON.parse((event as MessageEvent).data);
        const newStatus = data.status as PaymentStatus;
        setStatus(newStatus);
        if (isTerminal(newStatus)) {
          cleanup();
        }
      } catch {
        // Parse error
      }
    });

    es.onerror = () => {
      es.close();
      eventSourceRef.current = null;
      sseRetriesRef.current++;

      if (sseRetriesRef.current < MAX_SSE_RETRIES) {
        // Retry SSE after a brief delay
        setTimeout(startSSE, 1000);
      } else {
        // Fall back to polling
        startPolling();
      }
    };
  }, [orderId, accessToken, cleanup, isTerminal, startPolling]);

  useEffect(() => {
    if (!enabled || !swipePaymentId) return;

    sseRetriesRef.current = 0;
    startSSE();

    return cleanup;
  }, [enabled, swipePaymentId, startSSE, cleanup]);

  return { status, isPolling, error };
}
