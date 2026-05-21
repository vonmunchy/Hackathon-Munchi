"use client";

import { useState, use } from "react";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import { ProductGrid } from "@/components/buyer/product-grid";

export default function StorefrontPage({
  params,
}: {
  params: Promise<{ storeSlug: string }>;
}) {
  const { storeSlug } = use(params);
  const data = useQuery(api.products.listByStoreSlugWithVariants, { storeSlug });
  const [selectedCategory, setSelectedCategory] = useState<string>("All");

  // Loading state
  if (data === undefined) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-amethyst-500 border-t-transparent" />
      </div>
    );
  }

  // Store not found
  if (data === null) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <div className="text-5xl">🏝</div>
        <h1 className="font-display text-xl font-semibold text-slate-800">
          Store not found
        </h1>
        <p className="text-slate-500">
          This storefront doesn&apos;t exist or may have been removed.
        </p>
      </div>
    );
  }

  const { store, products } = data;

  // Gather unique categories
  const categories = [
    "All",
    ...Array.from(
      new Set(
        products
          .map((p) => p.category)
          .filter((c): c is string => Boolean(c)),
      ),
    ),
  ];

  // Filter active products by category
  const activeProducts = products.filter((p) => p.status === "active");
  const filtered =
    selectedCategory === "All"
      ? activeProducts
      : activeProducts.filter((p) => p.category === selectedCategory);

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      {/* Store header */}
      <div className="mb-6 flex items-center gap-4">
        {store.logoUrl ? (
          <img
            src={store.logoUrl}
            alt={store.name}
            className="h-14 w-14 rounded-full object-cover border border-slate-200"
          />
        ) : (
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-amethyst-100 text-amethyst-600 font-display font-bold text-xl">
            {store.name.charAt(0)}
          </div>
        )}
        <div>
          <h1 className="font-display text-xl font-bold text-slate-900">
            {store.name}
          </h1>
          {store.description && (
            <p className="mt-0.5 text-sm text-slate-500">{store.description}</p>
          )}
        </div>
      </div>

      {/* Category tabs */}
      {categories.length > 1 && (
        <div className="mb-5 -mx-4 px-4 overflow-x-auto scrollbar-hide">
          <div className="flex gap-2 pb-1">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setSelectedCategory(cat)}
                className={`shrink-0 rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                  selectedCategory === cat
                    ? "bg-amethyst-500 text-white"
                    : "bg-slate-100 text-slate-600 hover:bg-slate-200"
                }`}
              >
                {cat}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Products */}
      <ProductGrid products={filtered} storeSlug={storeSlug} />
    </div>
  );
}
