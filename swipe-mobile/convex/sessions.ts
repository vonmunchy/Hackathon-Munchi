import { mutation, query } from "./_generated/server";
import { v } from "convex/values";

const SESSION_TTL_MS = 24 * 60 * 60 * 1000; // 24 hours

export const create = mutation({
  args: {
    token: v.string(),
    storeId: v.id("stores"),
  },
  handler: async (ctx, args) => {
    // Clean up any existing sessions for this token
    const existing = await ctx.db
      .query("sessions")
      .withIndex("by_token", (q) => q.eq("token", args.token))
      .unique();
    if (existing) {
      await ctx.db.delete(existing._id);
    }

    const now = Date.now();
    return await ctx.db.insert("sessions", {
      token: args.token,
      storeId: args.storeId,
      createdAt: now,
      expiresAt: now + SESSION_TTL_MS,
    });
  },
});

export const validate = query({
  args: { token: v.string() },
  handler: async (ctx, args) => {
    const session = await ctx.db
      .query("sessions")
      .withIndex("by_token", (q) => q.eq("token", args.token))
      .unique();

    if (!session || Date.now() > session.expiresAt) {
      return null;
    }

    const store = await ctx.db.get(session.storeId);
    if (!store) return null;

    return { storeId: session.storeId, storeSlug: store.slug, storeName: store.name };
  },
});

export const remove = mutation({
  args: { token: v.string() },
  handler: async (ctx, args) => {
    const session = await ctx.db
      .query("sessions")
      .withIndex("by_token", (q) => q.eq("token", args.token))
      .unique();
    if (session) {
      await ctx.db.delete(session._id);
    }
  },
});
