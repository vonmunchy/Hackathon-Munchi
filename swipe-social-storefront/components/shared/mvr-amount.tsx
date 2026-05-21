"use client";

import { formatMVR } from "@/lib/format";

interface MvrAmountProps {
  amount: number;
  className?: string;
}

export function MvrAmount({ amount, className = "" }: MvrAmountProps) {
  return (
    <span className={`font-mono ${className}`}>{formatMVR(amount)}</span>
  );
}
