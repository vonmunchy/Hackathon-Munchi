"use client";

import Link from "next/link";
import Image from "next/image";
import { MvrAmount } from "@/components/shared/mvr-amount";
import type { Id } from "@/convex/_generated/dataModel";

interface ProductCardProps {
  product: {
    _id: Id<"products">;
    name: string;
    basePrice: number;
    imageUrls: string[];
    category?: string;
    status: string;
  };
  totalStock: number;
  storeSlug: string;
}

export function ProductCard({ product, totalStock, storeSlug }: ProductCardProps) {
  const hasImage = product.imageUrls.length > 0;

  return (
    <Link
      href={`/shop/${storeSlug}/product/${product._id}`}
      className="group block rounded-xl overflow-hidden bg-white transition-all duration-200 active:scale-[0.97] md:hover:scale-[1.02] md:hover:shadow-md"
    >
      {/* Image */}
      <div className="relative aspect-square rounded-xl overflow-hidden bg-slate-100">
        {hasImage ? (
          <Image
            src={product.imageUrls[0]}
            alt={product.name}
            fill
            className="object-cover"
            sizes="(max-width: 768px) 50vw, (max-width: 1024px) 33vw, 25vw"
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

        {/* Stock badges */}
        {totalStock === 0 && (
          <div className="absolute top-2 right-2 rounded-full bg-slate-700 px-2.5 py-0.5 text-xs font-medium text-white">
            Out of Stock
          </div>
        )}
        {totalStock > 0 && totalStock <= 3 && (
          <div className="absolute top-2 right-2 rounded-full bg-ruby-500 px-2.5 py-0.5 text-xs font-medium text-white">
            Low Stock
          </div>
        )}
      </div>

      {/* Info */}
      <div className="px-1 pt-2.5 pb-1">
        <h3 className="text-sm font-medium text-slate-800 line-clamp-2 leading-snug">
          {product.name}
        </h3>
        <MvrAmount
          amount={product.basePrice}
          className="mt-1 block text-sm text-amethyst-600"
        />
      </div>
    </Link>
  );
}
