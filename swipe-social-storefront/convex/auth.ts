import { QueryCtx, MutationCtx } from "./_generated/server";
import { Id } from "./_generated/dataModel";

/**
 * Validates a session token and returns the storeId.
 * Throws if the session is invalid or expired.
 * Used by all seller-facing queries and mutations for multi-tenant isolation.
 */
export async function authenticateStore(
  ctx: QueryCtx | MutationCtx,
  sessionToken: string,
): Promise<Id<"stores">> {
  const session = await ctx.db
    .query("sessions")
    .withIndex("by_token", (q) => q.eq("token", sessionToken))
    .unique();

  if (!session) {
    throw new Error("Invalid session");
  }

  if (Date.now() > session.expiresAt) {
    throw new Error("Session expired");
  }

  return session.storeId;
}

/**
 * Validates that a product belongs to the authenticated store.
 */
export async function validateProductOwnership(
  ctx: QueryCtx | MutationCtx,
  productId: Id<"products">,
  storeId: Id<"stores">,
): Promise<void> {
  const product = await ctx.db.get(productId);
  if (!product || product.storeId !== storeId) {
    throw new Error("Product not found or access denied");
  }
}

/**
 * Validates that an order belongs to the authenticated store.
 */
export async function validateOrderOwnership(
  ctx: QueryCtx | MutationCtx,
  orderId: Id<"orders">,
  storeId: Id<"stores">,
): Promise<void> {
  const order = await ctx.db.get(orderId);
  if (!order || order.storeId !== storeId) {
    throw new Error("Order not found or access denied");
  }
}

/**
 * Validates that a variant belongs to a product owned by the authenticated store.
 */
export async function validateVariantOwnership(
  ctx: QueryCtx | MutationCtx,
  variantId: Id<"productVariants">,
  storeId: Id<"stores">,
): Promise<void> {
  const variant = await ctx.db.get(variantId);
  if (!variant) throw new Error("Variant not found");
  const product = await ctx.db.get(variant.productId);
  if (!product || product.storeId !== storeId) {
    throw new Error("Variant not found or access denied");
  }
}
