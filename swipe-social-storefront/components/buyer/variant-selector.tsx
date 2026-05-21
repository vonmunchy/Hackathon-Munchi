"use client";

import type { Id } from "@/convex/_generated/dataModel";

interface Variant {
  _id: Id<"productVariants">;
  variantName: string;
  stockAvailable: number;
  priceOverride?: number;
}

interface VariantSelectorProps {
  variants: Variant[];
  selectedId: Id<"productVariants"> | null;
  onSelect: (id: Id<"productVariants">) => void;
}

export function VariantSelector({
  variants,
  selectedId,
  onSelect,
}: VariantSelectorProps) {
  const selected = variants.find((v) => v._id === selectedId);

  return (
    <div>
      <div className="flex flex-wrap gap-2">
        {variants.map((variant) => {
          const isSelected = variant._id === selectedId;
          const outOfStock = variant.stockAvailable === 0;

          return (
            <button
              key={variant._id}
              onClick={() => onSelect(variant._id)}
              disabled={outOfStock}
              className={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                isSelected
                  ? "bg-amethyst-500 text-white"
                  : outOfStock
                    ? "border border-slate-200 bg-slate-50 text-slate-300 line-through"
                    : "border border-slate-200 bg-slate-100 text-slate-700 hover:border-amethyst-300"
              }`}
            >
              {variant.variantName}
              {outOfStock && !isSelected && (
                <span className="ml-1 text-xs">Out</span>
              )}
            </button>
          );
        })}
      </div>

      {selected && (
        <p className="mt-2 text-sm text-slate-500">
          {selected.stockAvailable} available
        </p>
      )}
    </div>
  );
}
