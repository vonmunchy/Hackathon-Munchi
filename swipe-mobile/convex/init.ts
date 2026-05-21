import { internalMutation } from "./_generated/server";
import { v } from "convex/values";

export const seed = internalMutation({
  args: {},
  handler: async (ctx) => {
    // Idempotent: check if store already exists
    const existing = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", "island-finds-mv"))
      .unique();

    if (existing) {
      return;
    }

    // Create the store
    const storeId = await ctx.db.insert("stores", {
      name: "Island Finds MV",
      slug: "island-finds-mv",
      pin: "1234",
      description:
        "Curated finds from the Maldives — fashion, gifts, and handmade treasures",
      instagramUsername: "islandfinds.mv",
      facebookPageUrl: "https://facebook.com/islandfinds.mv",
      whatsappNumber: "+9607771234",
    });

    const now = Date.now();

    // Product 1: Black Abaya
    const abayaId = await ctx.db.insert("products", {
      storeId,
      name: "Black Abaya",
      description:
        "Elegant black abaya with intricate embroidery details. Perfect for special occasions and everyday grace.",
      basePrice: 650,
      currency: "MVR",
      category: "Fashion",
      imageUrls: ["/demo-products/abaya.png"],
      status: "active",
      createdAt: now,
    });
    await ctx.db.insert("productVariants", {
      productId: abayaId,
      variantName: "S",
      size: "S",
      stockAvailable: 2,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: abayaId,
      variantName: "M",
      size: "M",
      stockAvailable: 3,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: abayaId,
      variantName: "L",
      size: "L",
      stockAvailable: 1,
      stockSold: 0,
    });

    // Product 2: Eid Gift Box
    const giftBoxId = await ctx.db.insert("products", {
      storeId,
      name: "Eid Gift Box",
      description:
        "Beautifully curated gift box with premium Maldivian treats and accessories. Ideal for Eid celebrations.",
      basePrice: 450,
      currency: "MVR",
      category: "Gifts",
      imageUrls: ["/demo-products/gift-box.png"],
      status: "active",
      createdAt: now,
    });
    await ctx.db.insert("productVariants", {
      productId: giftBoxId,
      variantName: "Standard",
      stockAvailable: 8,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: giftBoxId,
      variantName: "Premium",
      priceOverride: 650,
      stockAvailable: 4,
      stockSold: 0,
    });

    // Product 3: iPhone Case — Maldives Edition
    const phoneCaseId = await ctx.db.insert("products", {
      storeId,
      name: "iPhone Case — Maldives Edition",
      description:
        "Protective phone case featuring a sleek matte finish with MagSafe support. Slim fit, premium feel.",
      basePrice: 120,
      currency: "MVR",
      category: "Accessories",
      imageUrls: ["/demo-products/phone-case.png"],
      status: "active",
      createdAt: now,
    });
    await ctx.db.insert("productVariants", {
      productId: phoneCaseId,
      variantName: "iPhone 15",
      stockAvailable: 12,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: phoneCaseId,
      variantName: "iPhone 14",
      stockAvailable: 5,
      stockSold: 0,
    });

    // Product 4: Handmade Coral Bracelet
    const braceletId = await ctx.db.insert("products", {
      storeId,
      name: "Handmade Coral Bracelet",
      description:
        "Artisan-crafted bracelet set inspired by Maldivian coral reefs. Each piece is unique and handmade with care.",
      basePrice: 85,
      currency: "MVR",
      category: "Accessories",
      imageUrls: ["/demo-products/bracelet.png"],
      status: "active",
      createdAt: now,
    });
    await ctx.db.insert("productVariants", {
      productId: braceletId,
      variantName: "Small",
      size: "Small",
      stockAvailable: 6,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: braceletId,
      variantName: "Medium",
      size: "Medium",
      stockAvailable: 4,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: braceletId,
      variantName: "Large",
      size: "Large",
      stockAvailable: 2,
      stockSold: 0,
    });

    // Product 5: Premium Dates Box
    const datesBoxId = await ctx.db.insert("products", {
      storeId,
      name: "Premium Dates Box",
      description:
        "Premium Ajwa dates sourced and packed fresh. A traditional gift of sweetness and hospitality.",
      basePrice: 280,
      currency: "MVR",
      category: "Food/Gifts",
      imageUrls: ["/demo-products/dates-box.png"],
      status: "active",
      createdAt: now,
    });
    await ctx.db.insert("productVariants", {
      productId: datesBoxId,
      variantName: "500g",
      stockAvailable: 10,
      stockSold: 0,
    });
    await ctx.db.insert("productVariants", {
      productId: datesBoxId,
      variantName: "1kg",
      priceOverride: 480,
      stockAvailable: 3,
      stockSold: 0,
    });
  },
});
