import { query, mutation, internalMutation } from "./_generated/server";
import { internal } from "./_generated/api";
import { v } from "convex/values";
import { authenticateStore } from "./auth";

export const getById = query({
  args: { storeId: v.id("stores") },
  handler: async (ctx, args) => {
    return await ctx.db.get(args.storeId);
  },
});

export const getBySlug = query({
  args: { slug: v.string() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.slug))
      .unique();
  },
});

export const verifyPin = query({
  args: { slug: v.string(), pin: v.string() },
  handler: async (ctx, args) => {
    const store = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.slug))
      .unique();

    if (!store) {
      return { found: false as const };
    }

    if (store.pin !== args.pin) {
      return { found: true as const, pinMatch: false as const };
    }

    return { found: true as const, pinMatch: true as const, storeId: store._id };
  },
});

// One-time admin mutation to link Facebook account to existing store and clean up orphan
export const linkFacebookAndCleanup = mutation({
  args: {
    targetStoreSlug: v.string(),
    orphanStoreSlug: v.string(),
    facebookUserId: v.string(),
    facebookPageId: v.string(),
  },
  handler: async (ctx, args) => {
    const target = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.targetStoreSlug))
      .unique();
    if (!target) throw new Error("Target store not found");

    const orphan = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.orphanStoreSlug))
      .unique();

    // Link Facebook to real store
    await ctx.db.patch(target._id, {
      facebookUserId: args.facebookUserId,
      facebookPageId: args.facebookPageId,
    });

    // Delete orphan
    if (orphan) {
      await ctx.db.delete(orphan._id);
    }

    return { linked: target._id, deleted: orphan?._id };
  },
});

export const updateProfile = internalMutation({
  args: {
    storeId: v.id("stores"),
    description: v.optional(v.string()),
    instagramUsername: v.optional(v.string()),
    facebookPageUrl: v.optional(v.string()),
    whatsappNumber: v.optional(v.string()),
    logoUrl: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const { storeId, ...fields } = args;
    await ctx.db.patch(storeId, fields);
  },
});

export const findOrCreateByFacebook = mutation({
  args: {
    facebookUserId: v.string(),
    facebookPageId: v.optional(v.string()),
    name: v.string(),
  },
  handler: async (ctx, args) => {
    // Priority 1: Try to find by Facebook Page ID (the business identity)
    if (args.facebookPageId) {
      const allStores = await ctx.db.query("stores").collect();
      const byPageId = allStores.find((s) => s.facebookPageId === args.facebookPageId);
      if (byPageId) {
        await ctx.db.patch(byPageId._id, { facebookUserId: args.facebookUserId });
        return { storeId: byPageId._id, slug: byPageId.slug, name: byPageId.name };
      }
    }

    // Priority 2: Try to match by name/slug derived from page name
    const slug = args.name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "") || `store-${args.facebookUserId}`;

    const bySlug = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", slug))
      .unique();

    if (bySlug) {
      // Link existing store to this Facebook account
      await ctx.db.patch(bySlug._id, {
        facebookUserId: args.facebookUserId,
        facebookPageId: args.facebookPageId,
      });
      return { storeId: bySlug._id, slug: bySlug.slug, name: bySlug.name };
    }

    // Create new store only if no match found
    const storeId = await ctx.db.insert("stores", {
      name: args.name,
      slug,
      pin: "0000",
      facebookUserId: args.facebookUserId,
      facebookPageId: args.facebookPageId,
    });

    return { storeId, slug, name: args.name };
  },
});

export const findOrCreateByInstagram = mutation({
  args: {
    instagramUserId: v.string(),
    instagramUsername: v.optional(v.string()),
    name: v.string(),
  },
  handler: async (ctx, args) => {
    // Try to find existing store by Instagram user ID
    const existing = await ctx.db
      .query("stores")
      .withIndex("by_instagramUserId", (q) =>
        q.eq("instagramUserId", args.instagramUserId),
      )
      .unique();

    if (existing) {
      if (args.instagramUsername && existing.instagramUsername !== args.instagramUsername) {
        await ctx.db.patch(existing._id, { instagramUsername: args.instagramUsername });
      }
      return { storeId: existing._id, slug: existing.slug, name: existing.name };
    }

    // Create new store
    const slug = (args.instagramUsername || args.name)
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "") || `store-${args.instagramUserId}`;

    const storeId = await ctx.db.insert("stores", {
      name: args.name,
      slug,
      pin: "0000",
      instagramUserId: args.instagramUserId,
      instagramUsername: args.instagramUsername,
    });

    return { storeId, slug, name: args.name };
  },
});

export const setSwipeCredentials = mutation({
  args: {
    sessionToken: v.string(),
    swipeClientId: v.string(),
    swipeClientSecret: v.string(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await ctx.db.patch(storeId, {
      swipeClientId: args.swipeClientId,
      swipeClientSecret: args.swipeClientSecret,
    });
  },
});

export const setContactPhone = mutation({
  args: {
    sessionToken: v.string(),
    contactPhone: v.string(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await ctx.db.patch(storeId, {
      contactPhone: args.contactPhone,
      whatsappNumber: args.contactPhone,
    });
  },
});

export const completeOnboarding = mutation({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await ctx.db.patch(storeId, { onboardingComplete: true });
  },
});

export const setSellerType = mutation({
  args: {
    sessionToken: v.string(),
    sellerType: v.string(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    await ctx.db.patch(storeId, { sellerType: args.sellerType });
    return { success: true };
  },
});

export const setCryptoWallet = mutation({
  args: {
    sessionToken: v.string(),
    cryptoWalletAddress: v.string(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    const addr = args.cryptoWalletAddress;
    if (!addr.startsWith("T") || addr.length !== 34 || !/^[A-Za-z0-9]+$/.test(addr)) {
      throw new Error("Invalid TRC20 wallet address. Must start with T and be exactly 34 alphanumeric characters.");
    }
    await ctx.db.patch(storeId, { cryptoWalletAddress: addr });
    return { success: true };
  },
});

export const updateStoreProfile = mutation({
  args: {
    sessionToken: v.string(),
    description: v.optional(v.string()),
    instagramUsername: v.optional(v.string()),
    facebookPageUrl: v.optional(v.string()),
    whatsappNumber: v.optional(v.string()),
    logoUrl: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    const { sessionToken: _, ...fields } = args;
    await ctx.db.patch(storeId, fields);
  },
});

// ---------- Admin ----------

export const listAll = query({
  args: {},
  handler: async (ctx) => {
    const stores = await ctx.db.query("stores").collect();
    return stores.map((s) => ({
      _id: s._id,
      name: s.name,
      slug: s.slug,
      sellerType: s.sellerType ?? null,
      onboardingComplete: s.onboardingComplete ?? false,
      hasSwipeCredentials: !!(s.swipeClientId && s.swipeClientSecret),
      contactPhone: s.contactPhone ?? null,
    }));
  },
});

export const resetStore = mutation({
  args: { storeId: v.id("stores") },
  handler: async (ctx, args) => {
    const store = await ctx.db.get(args.storeId);
    if (!store) {
      throw new Error("Store not found");
    }
    await ctx.db.patch(args.storeId, {
      onboardingComplete: undefined,
      swipeClientId: undefined,
      swipeClientSecret: undefined,
    });
    await ctx.runMutation(internal.sessions.deleteByStoreId, {
      storeId: args.storeId,
    });
    return { success: true, storeName: store.name };
  },
});
