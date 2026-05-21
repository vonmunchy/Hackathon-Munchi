import { mutation } from "./_generated/server";
import { v } from "convex/values";

// Helper: generate random alphanumeric string
function randomAlphanumeric(length: number): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Helper: generate random hex string
function randomHex(length: number): string {
  const chars = "0123456789abcdef";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * Seed dummy exchange listings and transactions for demo/testing.
 * This is a public mutation (no auth) intended only for dev/hackathon use.
 */
export const seedExchangeData = mutation({
  args: {
    storeSlug: v.string(),
  },
  handler: async (ctx, args) => {
    // 1. Find the store by slug
    const store = await ctx.db
      .query("stores")
      .withIndex("by_slug", (q) => q.eq("slug", args.storeSlug))
      .first();

    if (!store) {
      throw new Error(`Store with slug "${args.storeSlug}" not found`);
    }

    const storeId = store._id;

    // 2. Patch store to have sellerType "both" and a crypto wallet if not set
    if (!store.sellerType) {
      await ctx.db.patch(storeId, {
        sellerType: "both",
        cryptoWalletAddress: "TRX7k9mVHcFz5NwPqJ3bY8sLdA2eG4xK5n",
      });
    }

    // 3. Check if listings already exist for this store to avoid duplicates
    const existing = await ctx.db
      .query("exchangeListings")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .first();

    if (existing) {
      return {
        message: "Exchange data already seeded for this store",
        created: 0,
      };
    }

    const now = Date.now();
    let created = 0;

    // 4. Create listing 1: 100 USDT at 25.50
    await ctx.db.insert("exchangeListings", {
      storeId,
      usdtAmount: 100,
      rate: 25.5,
      availableBalance: 100,
      reservedBalance: 0,
      partialAllowed: true,
      walletAddress: "T" + randomAlphanumeric(33),
      status: "active",
      createdAt: now - 3600000, // 1 hour ago
    });
    created++;

    // 5. Create listing 2: 250 USDT at 25.00
    await ctx.db.insert("exchangeListings", {
      storeId,
      usdtAmount: 250,
      rate: 25.0,
      availableBalance: 250,
      reservedBalance: 0,
      partialAllowed: true,
      walletAddress: "T" + randomAlphanumeric(33),
      status: "active",
      createdAt: now - 7200000, // 2 hours ago
    });
    created++;

    // 6. Create listing 3: 50 USDT at 26.00 (no partial)
    await ctx.db.insert("exchangeListings", {
      storeId,
      usdtAmount: 50,
      rate: 26.0,
      availableBalance: 50,
      reservedBalance: 0,
      partialAllowed: false,
      walletAddress: "T" + randomAlphanumeric(33),
      status: "active",
      createdAt: now - 1800000, // 30 min ago
    });
    created++;

    // 7. Create listing 4: 500 USDT at 24.80 (partially sold - 150 gone)
    const listing4Id = await ctx.db.insert("exchangeListings", {
      storeId,
      usdtAmount: 500,
      rate: 24.8,
      availableBalance: 350,
      reservedBalance: 0,
      partialAllowed: true,
      walletAddress: "T" + randomAlphanumeric(33),
      status: "active",
      createdAt: now - 86400000, // 1 day ago
    });
    created++;

    // 8. Create 2 completed transactions for listing 4
    await ctx.db.insert("exchangeTransactions", {
      listingId: listing4Id,
      buyerWallet: "TBuyer1" + randomAlphanumeric(27),
      usdtAmount: 100,
      mvrAmount: 100 * 24.8, // 2480 MVR
      txHash: "0x" + randomHex(64),
      status: "completed",
      swipePaymentId: "swipe_demo_" + randomAlphanumeric(8),
      createdAt: now - 43200000, // 12 hours ago
    });
    created++;

    await ctx.db.insert("exchangeTransactions", {
      listingId: listing4Id,
      buyerWallet: "TBuyer2" + randomAlphanumeric(27),
      usdtAmount: 50,
      mvrAmount: 50 * 24.8, // 1240 MVR
      txHash: "0x" + randomHex(64),
      status: "completed",
      swipePaymentId: "swipe_demo_" + randomAlphanumeric(8),
      createdAt: now - 21600000, // 6 hours ago
    });
    created++;

    return {
      message: `Seeded ${created} exchange items (4 listings + 2 transactions)`,
      created,
    };
  },
});
