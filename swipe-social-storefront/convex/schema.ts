import { defineSchema, defineTable } from "convex/server";
import { v } from "convex/values";

export default defineSchema({
  stores: defineTable({
    name: v.string(),
    slug: v.string(),
    pin: v.string(),
    description: v.optional(v.string()),
    instagramUsername: v.optional(v.string()),
    facebookPageUrl: v.optional(v.string()),
    whatsappNumber: v.optional(v.string()),
    logoUrl: v.optional(v.string()),
    facebookUserId: v.optional(v.string()),
    facebookPageId: v.optional(v.string()),
    instagramUserId: v.optional(v.string()),
    swipeClientId: v.optional(v.string()),
    swipeClientSecret: v.optional(v.string()),
    onboardingComplete: v.optional(v.boolean()),
    contactPhone: v.optional(v.string()),
    sellerType: v.optional(v.string()),
    cryptoWalletAddress: v.optional(v.string()),
  })
    .index("by_slug", ["slug"])
    .index("by_facebookUserId", ["facebookUserId"])
    .index("by_instagramUserId", ["instagramUserId"]),

  products: defineTable({
    storeId: v.id("stores"),
    name: v.string(),
    description: v.optional(v.string()),
    basePrice: v.number(),
    currency: v.string(),
    category: v.optional(v.string()),
    imageUrls: v.array(v.string()),
    status: v.string(),
    createdAt: v.number(),
  }).index("by_storeId", ["storeId"]),

  productVariants: defineTable({
    productId: v.id("products"),
    variantName: v.string(),
    size: v.optional(v.string()),
    color: v.optional(v.string()),
    priceOverride: v.optional(v.number()),
    stockAvailable: v.number(),
    stockSold: v.number(),
  }).index("by_productId", ["productId"]),

  orders: defineTable({
    storeId: v.id("stores"),
    productId: v.id("products"),
    variantId: v.id("productVariants"),
    quantity: v.number(),
    unitPrice: v.number(),
    totalAmount: v.number(),
    currency: v.string(),
    customerName: v.string(),
    customerPhone: v.string(),
    deliveryLocation: v.string(),
    deliveryAddress: v.string(),
    deliveryTimePreference: v.optional(v.string()),
    status: v.string(),
    paymentStatus: v.string(),
    swipePaymentId: v.optional(v.string()),
    swipeReference: v.optional(v.string()),
    swipeShortCode: v.optional(v.string()),
    swipeQrData: v.optional(v.string()),
    swipePaymentUrl: v.optional(v.string()),
    accessToken: v.string(),
    createdAt: v.number(),
    updatedAt: v.number(),
  })
    .index("by_storeId", ["storeId"])
    .index("by_storeId_and_status", ["storeId", "status"])
    .index("by_accessToken", ["accessToken"]),

  sessions: defineTable({
    token: v.string(),
    storeId: v.id("stores"),
    createdAt: v.number(),
    expiresAt: v.number(),
  })
    .index("by_token", ["token"])
    .index("by_storeId", ["storeId"]),

  exchangeListings: defineTable({
    storeId: v.id("stores"),
    usdtAmount: v.number(),
    rate: v.number(),
    availableBalance: v.number(),
    reservedBalance: v.number(),
    partialAllowed: v.boolean(),
    walletAddress: v.string(),
    status: v.string(),
    createdAt: v.number(),
  })
    .index("by_status", ["status"])
    .index("by_storeId", ["storeId"]),

  exchangeTransactions: defineTable({
    listingId: v.id("exchangeListings"),
    buyerWallet: v.string(),
    usdtAmount: v.number(),
    mvrAmount: v.number(),
    txHash: v.string(),
    status: v.string(),
    swipePaymentId: v.optional(v.string()),
    createdAt: v.number(),
  })
    .index("by_listingId", ["listingId"]),

  exchangeReservations: defineTable({
    listingId: v.id("exchangeListings"),
    amount: v.number(),
    buyerWallet: v.string(),
    expiresAt: v.number(),
    status: v.string(),
  })
    .index("by_listingId", ["listingId"])
    .index("by_status", ["status"]),
});
