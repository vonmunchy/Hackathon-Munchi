import { query, mutation, internalMutation } from "./_generated/server";
import { v } from "convex/values";
import { authenticateStore } from "./auth";

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

// 1. createListing — authenticated seller creates a new listing
export const createListing = mutation({
  args: {
    sessionToken: v.string(),
    usdtAmount: v.number(),
    rate: v.number(),
    partialAllowed: v.boolean(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);

    // Generate simulated TRC20 wallet address
    const walletAddress = "T" + randomAlphanumeric(33);

    const listingId = await ctx.db.insert("exchangeListings", {
      storeId,
      usdtAmount: args.usdtAmount,
      rate: args.rate,
      availableBalance: 0,
      reservedBalance: 0,
      partialAllowed: args.partialAllowed,
      walletAddress,
      status: "pending_deposit",
      createdAt: Date.now(),
    });

    return { listingId, walletAddress };
  },
});

// 2. confirmDeposit — seller confirms simulated deposit
export const confirmDeposit = mutation({
  args: {
    sessionToken: v.string(),
    listingId: v.id("exchangeListings"),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);

    const listing = await ctx.db.get(args.listingId);
    if (!listing || listing.storeId !== storeId) {
      throw new Error("Listing not found or access denied");
    }
    if (listing.status !== "pending_deposit") {
      throw new Error("Listing is not pending deposit");
    }

    await ctx.db.patch(args.listingId, {
      status: "active",
      availableBalance: listing.usdtAmount,
    });

    return { success: true as const };
  },
});

// 3. createReservation — buyer reserves USDT amount
export const createReservation = mutation({
  args: {
    listingId: v.id("exchangeListings"),
    amount: v.number(),
    buyerWallet: v.string(),
  },
  handler: async (ctx, args) => {
    // Validate TRC20 wallet format
    if (!args.buyerWallet.startsWith("T") || args.buyerWallet.length !== 34) {
      throw new Error("Invalid TRC20 wallet address: must start with T and be 34 characters");
    }

    const listing = await ctx.db.get(args.listingId);
    if (!listing) {
      throw new Error("Listing not found");
    }
    if (listing.status !== "active") {
      throw new Error("Listing is not active");
    }
    if (args.amount > listing.availableBalance) {
      throw new Error("Insufficient available balance");
    }

    // Update listing balances
    await ctx.db.patch(args.listingId, {
      availableBalance: listing.availableBalance - args.amount,
      reservedBalance: listing.reservedBalance + args.amount,
    });

    // Create reservation with 5-minute expiry
    const reservationId = await ctx.db.insert("exchangeReservations", {
      listingId: args.listingId,
      amount: args.amount,
      buyerWallet: args.buyerWallet,
      expiresAt: Date.now() + 5 * 60 * 1000,
      status: "active",
    });

    return { reservationId };
  },
});

// 4. confirmPurchase — completes purchase after Swipe payment
export const confirmPurchase = mutation({
  args: {
    reservationId: v.id("exchangeReservations"),
    swipePaymentId: v.string(),
  },
  handler: async (ctx, args) => {
    const reservation = await ctx.db.get(args.reservationId);
    if (!reservation) {
      throw new Error("Reservation not found");
    }
    if (reservation.status !== "active") {
      throw new Error("Reservation is not active");
    }
    if (Date.now() > reservation.expiresAt) {
      throw new Error("Reservation has expired");
    }

    // Re-read listing for consistency
    const listing = await ctx.db.get(reservation.listingId);
    if (!listing) {
      throw new Error("Listing not found");
    }

    // Generate simulated tx hash
    const txHash = "0x" + randomHex(64);

    // Calculate MVR amount based on rate
    const mvrAmount = reservation.amount * listing.rate;

    // Create transaction record
    await ctx.db.insert("exchangeTransactions", {
      listingId: reservation.listingId,
      buyerWallet: reservation.buyerWallet,
      usdtAmount: reservation.amount,
      mvrAmount,
      txHash,
      status: "completed",
      swipePaymentId: args.swipePaymentId,
      createdAt: Date.now(),
    });

    // Decrement reserved balance on listing
    const newReservedBalance = listing.reservedBalance - reservation.amount;
    const updates: Record<string, unknown> = {
      reservedBalance: newReservedBalance,
    };

    // If both balances are zero, mark as sold out
    if (listing.availableBalance + newReservedBalance === 0) {
      updates.status = "sold_out";
    }

    await ctx.db.patch(reservation.listingId, updates);

    // Mark reservation as completed
    await ctx.db.patch(args.reservationId, {
      status: "completed",
    });

    return { txHash, usdtAmount: reservation.amount, mvrAmount };
  },
});

// 5. expireReservation — called by scheduler
export const expireReservation = internalMutation({
  args: {
    reservationId: v.id("exchangeReservations"),
  },
  handler: async (ctx, args) => {
    const reservation = await ctx.db.get(args.reservationId);
    if (!reservation) return;
    if (reservation.status !== "active") return;
    if (Date.now() <= reservation.expiresAt) return;

    // Restore available balance on listing
    const listing = await ctx.db.get(reservation.listingId);
    if (listing) {
      await ctx.db.patch(reservation.listingId, {
        availableBalance: listing.availableBalance + reservation.amount,
        reservedBalance: listing.reservedBalance - reservation.amount,
      });
    }

    await ctx.db.patch(args.reservationId, {
      status: "expired",
    });
  },
});

// 6. withdrawListing — seller withdraws unsold USDT
export const withdrawListing = mutation({
  args: {
    sessionToken: v.string(),
    listingId: v.id("exchangeListings"),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);

    const listing = await ctx.db.get(args.listingId);
    if (!listing || listing.storeId !== storeId) {
      throw new Error("Listing not found or access denied");
    }
    if (listing.reservedBalance !== 0) {
      throw new Error("Cannot withdraw while there are active reservations");
    }

    const amountWithdrawn = listing.availableBalance;

    await ctx.db.patch(args.listingId, {
      status: "withdrawn",
      availableBalance: 0,
    });

    return { success: true as const, amountWithdrawn };
  },
});

// 7. getActiveListings — public, no auth
export const getActiveListings = query({
  args: {},
  handler: async (ctx) => {
    const listings = await ctx.db
      .query("exchangeListings")
      .withIndex("by_status", (q) => q.eq("status", "active"))
      .collect();

    // Filter for available balance > 0 and join with store name
    const results = [];
    for (const listing of listings) {
      if (listing.availableBalance <= 0) continue;
      const store = await ctx.db.get(listing.storeId);
      results.push({
        ...listing,
        storeName: store?.name ?? "Unknown",
      });
    }

    // Sort by rate ascending (lowest first)
    results.sort((a, b) => a.rate - b.rate);

    return results;
  },
});

// 8. getSellerListings — authenticated
export const getSellerListings = query({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);

    return await ctx.db
      .query("exchangeListings")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .order("desc")
      .collect();
  },
});

// 9. getListingDetails — public
export const getListingDetails = query({
  args: { listingId: v.id("exchangeListings") },
  handler: async (ctx, args) => {
    const listing = await ctx.db.get(args.listingId);
    if (!listing) return null;

    const store = await ctx.db.get(listing.storeId);

    return {
      ...listing,
      storeName: store?.name ?? "Unknown",
    };
  },
});
