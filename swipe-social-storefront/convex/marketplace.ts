import { query } from "./_generated/server";

export const getAllActiveProducts = query({
  args: {},
  handler: async (ctx) => {
    const products = await ctx.db.query("products").collect();
    const activeProducts = products.filter((p) => p.status === "active");

    // Sort by createdAt descending (newest first)
    activeProducts.sort((a, b) => b.createdAt - a.createdAt);

    // Join with store info
    const productsWithStore = await Promise.all(
      activeProducts.map(async (product) => {
        const store = await ctx.db.get(product.storeId);
        return {
          ...product,
          storeName: store?.name ?? "Unknown Seller",
          storeSlug: store?.slug ?? "",
        };
      }),
    );

    return productsWithStore;
  },
});
