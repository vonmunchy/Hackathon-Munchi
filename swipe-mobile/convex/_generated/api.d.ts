/* eslint-disable */
/**
 * Generated `api` utility.
 *
 * THIS CODE IS AUTOMATICALLY GENERATED.
 *
 * To regenerate, run `npx convex dev`.
 * @module
 */

import type * as auth from "../auth.js";
import type * as exchange from "../exchange.js";
import type * as init from "../init.js";
import type * as marketplace from "../marketplace.js";
import type * as orders from "../orders.js";
import type * as products from "../products.js";
import type * as reseed from "../reseed.js";
import type * as seedExchange from "../seedExchange.js";
import type * as sessions from "../sessions.js";
import type * as storage from "../storage.js";
import type * as stores from "../stores.js";

import type {
  ApiFromModules,
  FilterApi,
  FunctionReference,
} from "convex/server";

declare const fullApi: ApiFromModules<{
  auth: typeof auth;
  exchange: typeof exchange;
  init: typeof init;
  marketplace: typeof marketplace;
  orders: typeof orders;
  products: typeof products;
  reseed: typeof reseed;
  seedExchange: typeof seedExchange;
  sessions: typeof sessions;
  storage: typeof storage;
  stores: typeof stores;
}>;

/**
 * A utility for referencing Convex functions in your app's public API.
 *
 * Usage:
 * ```js
 * const myFunctionReference = api.myModule.myFunction;
 * ```
 */
export declare const api: FilterApi<
  typeof fullApi,
  FunctionReference<any, "public">
>;

/**
 * A utility for referencing Convex functions in your app's internal API.
 *
 * Usage:
 * ```js
 * const myFunctionReference = internal.myModule.myFunction;
 * ```
 */
export declare const internal: FilterApi<
  typeof fullApi,
  FunctionReference<any, "internal">
>;

export declare const components: {};
