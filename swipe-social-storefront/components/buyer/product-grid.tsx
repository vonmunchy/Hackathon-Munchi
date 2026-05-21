"use client";

import { ProductCard } from "./product-card";
import type { Id } from "@/convex/_generated/dataModel";

interface Variant {
  _id: Id<"productVariants">;
  stockAvailable: number;
}

interface Product {
  _id: Id<"products">;
  name: string;
  basePrice: number;
  imageUrls: string[];
  category?: string;
  status: string;
  variants: Variant[];
}

interface ProductGridProps {
  products: Product[];
  storeSlug: string;
}

export function ProductGrid({ products, storeSlug }: ProductGridProps) {
  if (products.length === 0) {
    return (
      <div className="py-16 text-center text-slate-400">
        <p className="text-lg">No products found</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4">
      {products.map((product) => {
        const totalStock = product.variants.reduce(
          (sum, v) => sum + v.stockAvailable,
          0,
        );
        return (
          <ProductCard
            key={product._id}
            product={product}
            totalStock={totalStock}
            storeSlug={storeSlug}
          />
        );
      })}
    </div>
  );
}
