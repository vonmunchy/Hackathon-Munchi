import { mutation, query, action } from "./_generated/server";
import { v } from "convex/values";
import { api } from "./_generated/api";

/**
 * Generate an upload URL for client-side file uploads.
 */
export const generateUploadUrl = mutation({
  handler: async (ctx) => {
    return await ctx.storage.generateUploadUrl();
  },
});

/**
 * Get the public URL for a stored file.
 */
export const getUrl = query({
  args: { storageId: v.id("_storage") },
  handler: async (ctx, args) => {
    return await ctx.storage.getUrl(args.storageId);
  },
});

/**
 * Store a file from a URL into Convex storage.
 */
export const storeFromUrl = action({
  args: { url: v.string(), contentType: v.string() },
  handler: async (ctx, args) => {
    const res = await fetch(args.url);
    if (!res.ok) throw new Error(`Failed to fetch ${args.url}: ${res.status}`);
    const blob = await res.blob();
    const storageId = await ctx.storage.store(blob);
    const publicUrl = await ctx.storage.getUrl(storageId);
    return { storageId, publicUrl };
  },
});

/**
 * Update a product's image URLs.
 */
export const updateProductImages = mutation({
  args: {
    productId: v.id("products"),
    imageUrls: v.array(v.string()),
  },
  handler: async (ctx, args) => {
    await ctx.db.patch(args.productId, { imageUrls: args.imageUrls });
  },
});
