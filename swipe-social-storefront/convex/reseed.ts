import { internalMutation } from "./_generated/server";
import { internal } from "./_generated/api";

export const run = internalMutation({
  args: {},
  handler: async (ctx) => {
    // Delete all existing data
    const stores = await ctx.db.query("stores").take(100);
    for (const store of stores) {
      await ctx.db.delete(store._id);
    }
    const products = await ctx.db.query("products").take(100);
    for (const product of products) {
      await ctx.db.delete(product._id);
    }
    const variants = await ctx.db.query("productVariants").take(200);
    for (const variant of variants) {
      await ctx.db.delete(variant._id);
    }
    const orders = await ctx.db.query("orders").take(200);
    for (const order of orders) {
      await ctx.db.delete(order._id);
    }

    // Re-seed
    await ctx.runMutation(internal.init.seed, {});
  },
});
