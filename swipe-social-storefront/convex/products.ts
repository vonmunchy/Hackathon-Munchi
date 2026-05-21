import { query, mutation } from "./_generated/server";
import { v } from "convex/values";
import { authenticateStore, validateProductOwnership, validateVariantOwnership } from "./auth";

export const listByStore = query({
  args: { storeId: v.id("stores") },
  handler: async (ctx, args) => {
    return await ctx.db
      .query("products")
      .withIndex("by_storeId", (q) => q.eq("storeId", args.storeId))
      .take(100);
  },
});

export const listByStoreSlug = query({
  args: { storeSlug: v.string() },
  handler: async (ctx, args) => {
    const store = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.storeSlug))
      .unique();

    if (!store) {
      return [];
    }

    return await ctx.db
      .query("products")
      .withIndex("by_storeId", (q) => q.eq("storeId", store._id))
      .take(100);
  },
});

export const listByStoreSlugWithVariants = query({
  args: { storeSlug: v.string() },
  handler: async (ctx, args) => {
    const store = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.storeSlug))
      .unique();

    if (!store) {
      return null;
    }

    const products = await ctx.db
      .query("products")
      .withIndex("by_storeId", (q) => q.eq("storeId", store._id))
      .take(100);

    const productsWithVariants = await Promise.all(
      products.map(async (product) => {
        const variants = await ctx.db
          .query("productVariants")
          .withIndex("by_productId", (q) => q.eq("productId", product._id))
          .take(50);
        return { ...product, variants };
      }),
    );

    return { store, products: productsWithVariants };
  },
});

export const getWithVariants = query({
  args: { productId: v.id("products") },
  handler: async (ctx, args) => {
    const product = await ctx.db.get(args.productId);
    if (!product) {
      return null;
    }

    const variants = await ctx.db
      .query("productVariants")
      .withIndex("by_productId", (q) => q.eq("productId", args.productId))
      .take(50);

    return { ...product, variants };
  },
});

export const create = mutation({
  args: {
    sessionToken: v.string(),
    name: v.string(),
    description: v.optional(v.string()),
    basePrice: v.number(),
    currency: v.string(),
    category: v.optional(v.string()),
    imageUrls: v.array(v.string()),
    status: v.string(),
    createdAt: v.number(),
    variants: v.array(
      v.object({
        variantName: v.string(),
        size: v.optional(v.string()),
        color: v.optional(v.string()),
        priceOverride: v.optional(v.number()),
        stockAvailable: v.number(),
        stockSold: v.number(),
      })
    ),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    const { variants, sessionToken: _, ...productFields } = args;
    const productId = await ctx.db.insert("products", { ...productFields, storeId });

    for (const variant of variants) {
      await ctx.db.insert("productVariants", {
        productId,
        ...variant,
      });
    }

    return productId;
  },
});

export const update = mutation({
  args: {
    sessionToken: v.string(),
    productId: v.id("products"),
    name: v.optional(v.string()),
    description: v.optional(v.string()),
    basePrice: v.optional(v.number()),
    currency: v.optional(v.string()),
    category: v.optional(v.string()),
    imageUrls: v.optional(v.array(v.string())),
    status: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await validateProductOwnership(ctx, args.productId, storeId);
    const { productId, sessionToken: _, ...fields } = args;
    await ctx.db.patch(productId, fields);
  },
});

export const archive = mutation({
  args: { sessionToken: v.string(), productId: v.id("products") },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await validateProductOwnership(ctx, args.productId, storeId);
    await ctx.db.patch(args.productId, { status: "archived" });
  },
});

export const addVariant = mutation({
  args: {
    sessionToken: v.string(),
    productId: v.id("products"),
    variantName: v.string(),
    size: v.optional(v.string()),
    color: v.optional(v.string()),
    priceOverride: v.optional(v.number()),
    stockAvailable: v.number(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await validateProductOwnership(ctx, args.productId, storeId);
    const { productId, sessionToken: _, ...fields } = args;
    return await ctx.db.insert("productVariants", {
      productId,
      ...fields,
      stockSold: 0,
    });
  },
});

export const updateVariant = mutation({
  args: {
    sessionToken: v.string(),
    variantId: v.id("productVariants"),
    variantName: v.optional(v.string()),
    size: v.optional(v.string()),
    color: v.optional(v.string()),
    priceOverride: v.optional(v.number()),
    stockAvailable: v.optional(v.number()),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await validateVariantOwnership(ctx, args.variantId, storeId);
    const { variantId, sessionToken: _, ...fields } = args;
    await ctx.db.patch(variantId, fields);
  },
});

export const removeVariant = mutation({
  args: { sessionToken: v.string(), variantId: v.id("productVariants") },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await validateVariantOwnership(ctx, args.variantId, storeId);
    await ctx.db.delete(args.variantId);
  },
});

export const importTestProducts = mutation({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    // Idempotency: skip if store already has products
    const existing = await ctx.db
      .query("products")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .take(1);
    if (existing.length > 0) {
      return { imported: 0, message: "Store already has products" };
    }

    const now = Date.now();
    const products = [
      {
        name: "Black Abaya",
        description: "Elegant black abaya with intricate embroidery details. Perfect for special occasions and everyday grace.",
        basePrice: 650,
        category: "Fashion",
        imageUrls: ["/demo-products/abaya.png"],
        variants: [
          { variantName: "S", size: "S", stockAvailable: 2 },
          { variantName: "M", size: "M", stockAvailable: 3 },
          { variantName: "L", size: "L", stockAvailable: 1 },
        ],
      },
      {
        name: "Eid Gift Box",
        description: "Beautifully curated gift box with premium Maldivian treats and accessories. Ideal for Eid celebrations.",
        basePrice: 450,
        category: "Gifts",
        imageUrls: ["/demo-products/gift-box.png"],
        variants: [
          { variantName: "Standard", stockAvailable: 8 },
          { variantName: "Premium", priceOverride: 650, stockAvailable: 4 },
        ],
      },
      {
        name: "iPhone Case — Maldives Edition",
        description: "Protective phone case featuring a sleek matte finish with MagSafe support. Slim fit, premium feel.",
        basePrice: 120,
        category: "Accessories",
        imageUrls: ["/demo-products/phone-case.png"],
        variants: [
          { variantName: "iPhone 15", stockAvailable: 12 },
          { variantName: "iPhone 14", stockAvailable: 5 },
        ],
      },
      {
        name: "Handmade Coral Bracelet",
        description: "Artisan-crafted bracelet set inspired by Maldivian coral reefs. Each piece is unique and handmade with care.",
        basePrice: 85,
        category: "Accessories",
        imageUrls: ["/demo-products/bracelet.png"],
        variants: [
          { variantName: "Small", size: "Small", stockAvailable: 6 },
          { variantName: "Medium", size: "Medium", stockAvailable: 4 },
          { variantName: "Large", size: "Large", stockAvailable: 2 },
        ],
      },
      {
        name: "Premium Dates Box",
        description: "Premium Ajwa dates sourced and packed fresh. A traditional gift of sweetness and hospitality.",
        basePrice: 280,
        category: "Food/Gifts",
        imageUrls: ["/demo-products/dates-box.png"],
        variants: [
          { variantName: "500g", stockAvailable: 10 },
          { variantName: "1kg", priceOverride: 480, stockAvailable: 3 },
        ],
      },
    ];

    for (const p of products) {
      const productId = await ctx.db.insert("products", {
        storeId,
        name: p.name,
        description: p.description,
        basePrice: p.basePrice,
        currency: "MVR",
        category: p.category,
        imageUrls: p.imageUrls,
        status: "active",
        createdAt: now,
      });

      for (const v of p.variants) {
        await ctx.db.insert("productVariants", {
          productId,
          variantName: v.variantName,
          size: "size" in v ? (v as { size: string }).size : undefined,
          priceOverride: "priceOverride" in v ? (v as { priceOverride: number }).priceOverride : undefined,
          stockAvailable: v.stockAvailable,
          stockSold: 0,
        });
      }
    }

    return { imported: products.length, message: "Test products imported" };
  },
});
