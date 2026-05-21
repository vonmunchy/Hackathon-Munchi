import { query, mutation, internalMutation } from "./_generated/server";
import { v } from "convex/values";
import { authenticateStore } from "./auth";

export const createOrder = mutation({
  args: {
    storeId: v.id("stores"),
    productId: v.id("products"),
    variantId: v.id("productVariants"),
    quantity: v.number(),
    customerName: v.string(),
    customerPhone: v.string(),
    deliveryLocation: v.string(),
    deliveryAddress: v.string(),
    deliveryTimePreference: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const variant = await ctx.db.get(args.variantId);
    if (!variant) {
      return { success: false as const, reason: "variant_not_found" };
    }

    // Advisory stock check (UX-only, not the atomic gate)
    if (variant.stockAvailable < args.quantity) {
      return { success: false as const, reason: "insufficient_stock" };
    }

    const product = await ctx.db.get(args.productId);
    if (!product) {
      return { success: false as const, reason: "product_not_found" };
    }

    const unitPrice = variant.priceOverride ?? product.basePrice;
    const totalAmount = unitPrice * args.quantity;
    const accessToken = crypto.randomUUID();
    const now = Date.now();

    const orderId = await ctx.db.insert("orders", {
      storeId: args.storeId,
      productId: args.productId,
      variantId: args.variantId,
      quantity: args.quantity,
      unitPrice,
      totalAmount,
      currency: product.currency,
      customerName: args.customerName,
      customerPhone: args.customerPhone,
      deliveryLocation: args.deliveryLocation,
      deliveryAddress: args.deliveryAddress,
      deliveryTimePreference: args.deliveryTimePreference,
      status: "pending",
      paymentStatus: "pending",
      accessToken,
      createdAt: now,
      updatedAt: now,
    });

    return { success: true as const, orderId, accessToken };
  },
});

export const getByIdWithToken = query({
  args: { orderId: v.id("orders"), accessToken: v.string() },
  handler: async (ctx, args) => {
    const order = await ctx.db.get(args.orderId);
    if (!order || order.accessToken !== args.accessToken) {
      return null;
    }
    return order;
  },
});

export const getByAccessToken = query({
  args: { accessToken: v.string() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query("orders")
      .withIndex("by_accessToken", (q) => q.eq("accessToken", args.accessToken))
      .unique();
  },
});

// Public version for ConvexHttpClient in API routes
export const confirmPaymentPublic = mutation({
  args: { orderId: v.id("orders") },
  handler: async (ctx, args) => {
    const order = await ctx.db.get(args.orderId);
    if (!order) {
      return { success: false as const, reason: "order_not_found" };
    }

    if (order.paymentStatus === "paid") {
      return { success: true as const };
    }

    const variant = await ctx.db.get(order.variantId);
    if (!variant) {
      return { success: false as const, reason: "variant_not_found" };
    }

    if (variant.stockAvailable < order.quantity) {
      await ctx.db.patch(args.orderId, {
        status: "cancelled",
        paymentStatus: "cancelled",
        updatedAt: Date.now(),
      });
      return { success: false as const, reason: "insufficient_stock" };
    }

    await ctx.db.patch(order.variantId, {
      stockAvailable: variant.stockAvailable - order.quantity,
      stockSold: variant.stockSold + order.quantity,
    });

    await ctx.db.patch(args.orderId, {
      status: "paid",
      paymentStatus: "paid",
      updatedAt: Date.now(),
    });

    return { success: true as const };
  },
});

export const confirmPayment = internalMutation({
  args: { orderId: v.id("orders") },
  handler: async (ctx, args) => {
    const order = await ctx.db.get(args.orderId);
    if (!order) {
      return { success: false as const, reason: "order_not_found" };
    }

    // Idempotent: already paid
    if (order.paymentStatus === "paid") {
      return { success: true as const };
    }

    const variant = await ctx.db.get(order.variantId);
    if (!variant) {
      return { success: false as const, reason: "variant_not_found" };
    }

    // Atomic stock gate
    if (variant.stockAvailable < order.quantity) {
      await ctx.db.patch(args.orderId, {
        status: "cancelled",
        paymentStatus: "cancelled",
        updatedAt: Date.now(),
      });
      return { success: false as const, reason: "insufficient_stock" };
    }

    // Atomically decrement stock and confirm order
    await ctx.db.patch(order.variantId, {
      stockAvailable: variant.stockAvailable - order.quantity,
      stockSold: variant.stockSold + order.quantity,
    });

    await ctx.db.patch(args.orderId, {
      status: "paid",
      paymentStatus: "paid",
      updatedAt: Date.now(),
    });

    return { success: true as const };
  },
});

export const updateSwipeData = internalMutation({
  args: {
    orderId: v.id("orders"),
    swipePaymentId: v.optional(v.string()),
    swipeReference: v.optional(v.string()),
    swipeShortCode: v.optional(v.string()),
    swipeQrData: v.optional(v.string()),
    swipePaymentUrl: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const { orderId, ...fields } = args;
    await ctx.db.patch(orderId, { ...fields, updatedAt: Date.now() });
  },
});

// Public version callable from ConvexHttpClient in Next.js API routes
export const setSwipeData = mutation({
  args: {
    orderId: v.id("orders"),
    swipePaymentId: v.optional(v.string()),
    swipeReference: v.optional(v.string()),
    swipeShortCode: v.optional(v.string()),
    swipeQrData: v.optional(v.string()),
    swipePaymentUrl: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const { orderId, ...fields } = args;
    await ctx.db.patch(orderId, { ...fields, updatedAt: Date.now() });
  },
});

export const updateOrderStatus = internalMutation({
  args: {
    orderId: v.id("orders"),
    newStatus: v.string(),
  },
  handler: async (ctx, args) => {
    const order = await ctx.db.get(args.orderId);
    if (!order) {
      return { success: false as const, reason: "order_not_found" };
    }

    // Validate transitions
    const validTransitions: Record<string, string[]> = {
      paid: ["shipped"],
      shipped: ["delivered"],
    };

    const allowed = validTransitions[order.status];
    if (!allowed || !allowed.includes(args.newStatus)) {
      return {
        success: false as const,
        reason: `invalid_transition_from_${order.status}_to_${args.newStatus}`,
      };
    }

    await ctx.db.patch(args.orderId, {
      status: args.newStatus,
      updatedAt: Date.now(),
    });

    return { success: true as const };
  },
});

export const fulfillOrder = mutation({
  args: {
    sessionToken: v.string(),
    orderId: v.id("orders"),
    newStatus: v.string(),
  },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    const order = await ctx.db.get(args.orderId);
    if (!order || order.storeId !== storeId) {
      return { success: false as const, reason: "order_not_found" };
    }

    const validTransitions: Record<string, string[]> = {
      paid: ["shipped"],
      shipped: ["delivered"],
    };

    const allowed = validTransitions[order.status];
    if (!allowed || !allowed.includes(args.newStatus)) {
      return {
        success: false as const,
        reason: `invalid_transition_from_${order.status}_to_${args.newStatus}`,
      };
    }

    await ctx.db.patch(args.orderId, {
      status: args.newStatus,
      updatedAt: Date.now(),
    });

    return { success: true as const };
  },
});

export const listByStore = query({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    return await ctx.db
      .query("orders")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .order("desc")
      .take(200);
  },
});

export const getOrderWithDetails = query({
  args: { orderId: v.id("orders") },
  handler: async (ctx, args) => {
    const order = await ctx.db.get(args.orderId);
    if (!order) {
      return null;
    }

    const product = await ctx.db.get(order.productId);
    const variant = await ctx.db.get(order.variantId);

    return { order, product, variant };
  },
});

export const getDashboardStats = query({
  args: { sessionToken: v.string() },
  handler: async (ctx, args) => {
    const storeId = await authenticateStore(ctx, args.sessionToken);
    const allOrders = await ctx.db
      .query("orders")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .take(1000);

    const now = Date.now();
    const startOfToday = now - (now % (24 * 60 * 60 * 1000));

    let totalSalesToday = 0;
    let totalSalesAllTime = 0;
    let paidCount = 0;
    let pendingCount = 0;
    let awaitingShipmentCount = 0;

    for (const order of allOrders) {
      if (order.paymentStatus === "paid") {
        paidCount++;
        totalSalesAllTime += order.totalAmount;
        if (order.createdAt >= startOfToday) {
          totalSalesToday += order.totalAmount;
        }
        if (order.status === "paid") {
          awaitingShipmentCount++;
        }
      }
      if (order.paymentStatus === "pending") {
        pendingCount++;
      }
    }

    // Count low stock variants across all products in the store
    const products = await ctx.db
      .query("products")
      .withIndex("by_storeId", (q) => q.eq("storeId", storeId))
      .take(100);

    let lowStockCount = 0;
    for (const product of products) {
      const variants = await ctx.db
        .query("productVariants")
        .withIndex("by_productId", (q) => q.eq("productId", product._id))
        .take(50);

      for (const variant of variants) {
        if (variant.stockAvailable <= 3) {
          lowStockCount++;
        }
      }
    }

    return {
      totalSalesToday,
      totalSalesAllTime,
      paidCount,
      pendingCount,
      lowStockCount,
      awaitingShipmentCount,
    };
  },
});
