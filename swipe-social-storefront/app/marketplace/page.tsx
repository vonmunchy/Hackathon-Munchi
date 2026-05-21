"use client";

import { useState, useMemo } from "react";
import { useQuery } from "convex/react";
import { api } from "@/convex/_generated/api";
import Link from "next/link";
import { MvrAmount } from "@/components/shared/mvr-amount";

export default function MarketplacePage() {
  const products = useQuery(api.marketplace.getAllActiveProducts);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [selectedSeller, setSelectedSeller] = useState<string | null>(null);

  // Extract unique categories and sellers
  const categories = useMemo(() => {
    if (!products) return [];
    const cats = new Set(
      products.map((p) => p.category).filter((c): c is string => Boolean(c)),
    );
    return Array.from(cats);
  }, [products]);

  const sellers = useMemo(() => {
    if (!products) return [];
    const s = new Set(products.map((p) => p.storeName));
    return Array.from(s);
  }, [products]);

  // Apply filters
  const filteredProducts = useMemo(() => {
    if (!products) return [];
    let result = products;

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          (p.description?.toLowerCase().includes(q) ?? false),
      );
    }

    if (selectedCategory) {
      result = result.filter((p) => p.category === selectedCategory);
    }

    if (selectedSeller) {
      result = result.filter((p) => p.storeName === selectedSeller);
    }

    return result;
  }, [products, searchQuery, selectedCategory, selectedSeller]);

  const hasActiveFilters = searchQuery || selectedCategory || selectedSeller;

  function clearFilters() {
    setSearchQuery("");
    setSelectedCategory(null);
    setSelectedSeller(null);
  }

  // Loading state
  if (products === undefined) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-6">
        <div className="mb-6">
          <div className="h-8 w-48 animate-pulse rounded bg-slate-200" />
          <div className="mt-2 h-5 w-72 animate-pulse rounded bg-slate-100" />
        </div>
        <div className="mb-5 h-10 w-full animate-pulse rounded-lg bg-slate-100" />
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="animate-pulse rounded-xl bg-white shadow-sm">
              <div className="aspect-square rounded-t-xl bg-slate-200" />
              <div className="p-3">
                <div className="h-4 w-3/4 rounded bg-slate-200" />
                <div className="mt-2 h-4 w-1/2 rounded bg-slate-100" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Empty state (no products at all)
  if (products.length === 0) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-4 text-center">
        <div className="text-5xl">🛍</div>
        <h1 className="font-display text-xl font-semibold text-slate-800">
          No products yet
        </h1>
        <p className="text-slate-500">
          Sellers haven&apos;t listed any products yet. Check back soon!
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="font-display text-2xl font-bold text-slate-900">
          Marketplace
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          Shop from local sellers, pay instantly with Swipe
        </p>
      </div>

      {/* Search bar */}
      <div className="relative mb-5 mx-auto w-full md:max-w-md md:mx-0">
        <svg
          className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={2}
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
          />
        </svg>
        <input
          type="text"
          placeholder="Search products..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full rounded-lg border border-slate-200 bg-white py-2.5 pl-10 pr-4 text-sm text-slate-800 shadow-sm placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
        />
      </div>

      {/* Category chips */}
      {categories.length > 0 && (
        <div className="mb-4 -mx-4 px-4 overflow-x-auto scrollbar-hide">
          <div className="flex flex-wrap gap-2 pb-1">
            <button
              onClick={() => setSelectedCategory(null)}
              className={`shrink-0 rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                selectedCategory === null
                  ? "bg-amethyst-500 text-white"
                  : "bg-slate-100 text-slate-600 hover:bg-slate-200"
              }`}
            >
              All
            </button>
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() =>
                  setSelectedCategory(selectedCategory === cat ? null : cat)
                }
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

      {/* Seller filter */}
      {sellers.length > 1 && (
        <div className="mb-5">
          <select
            value={selectedSeller ?? ""}
            onChange={(e) =>
              setSelectedSeller(e.target.value || null)
            }
            className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
          >
            <option value="">All Sellers</option>
            {sellers.map((seller) => (
              <option key={seller} value={seller}>
                {seller}
              </option>
            ))}
          </select>
        </div>
      )}

      {/* Active filters clear button */}
      {hasActiveFilters && (
        <div className="mb-4 flex items-center gap-2">
          <span className="text-sm text-slate-500">
            {filteredProducts.length} result{filteredProducts.length !== 1 ? "s" : ""}
          </span>
          <button
            onClick={clearFilters}
            className="text-sm font-medium text-amethyst-600 hover:text-amethyst-700"
          >
            Clear filters
          </button>
        </div>
      )}

      {/* No results state */}
      {filteredProducts.length === 0 && hasActiveFilters && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="text-4xl mb-3">🔍</div>
          <p className="text-lg font-medium text-slate-700">
            No products match your search
          </p>
          <button
            onClick={clearFilters}
            className="mt-3 text-sm font-medium text-amethyst-600 hover:text-amethyst-700"
          >
            Clear filters
          </button>
        </div>
      )}

      {/* Product grid */}
      {filteredProducts.length > 0 && (
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4">
          {filteredProducts.map((product) => (
            <Link
              key={product._id}
              href={`/shop/${product.storeSlug}/product/${product._id}`}
              className="group block rounded-xl overflow-hidden bg-white shadow-sm transition-all duration-200 active:scale-[0.97] md:hover:scale-[1.02] md:hover:shadow-md"
            >
              {/* Image */}
              <div className="relative aspect-square rounded-t-xl overflow-hidden bg-slate-100">
                {product.imageUrls.length > 0 ? (
                  <img
                    src={product.imageUrls[0]}
                    alt={product.name}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full items-center justify-center text-slate-400">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-10 w-10"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={1.5}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0 0 22.5 18.75V5.25A2.25 2.25 0 0 0 20.25 3H3.75A2.25 2.25 0 0 0 1.5 5.25v13.5A2.25 2.25 0 0 0 3.75 21Z"
                      />
                    </svg>
                  </div>
                )}
                {/* Category badge */}
                {product.category && (
                  <div className="absolute top-2 left-2 rounded-full bg-white/90 px-2 py-0.5 text-xs font-medium text-slate-700 backdrop-blur-sm">
                    {product.category}
                  </div>
                )}
              </div>

              {/* Info */}
              <div className="px-3 pt-2.5 pb-3">
                <h3 className="text-sm font-medium text-slate-800 line-clamp-2 leading-snug">
                  {product.name}
                </h3>
                <div className="mt-1.5 flex items-center gap-1 text-xs text-slate-500">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-3 w-3"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth={2}
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z"
                    />
                  </svg>
                  <span>{product.storeName}</span>
                </div>
                <MvrAmount
                  amount={product.basePrice}
                  className="mt-1.5 block text-sm font-bold text-amethyst-600"
                />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
