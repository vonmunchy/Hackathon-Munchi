import { SESSION_TTL_MS } from "./constants";

export interface MetaTokens {
  facebookPageAccessToken?: string;
  facebookPageId?: string;
  facebookPageName?: string;
  facebookUserId?: string;
  instagramAccessToken?: string;
  instagramUserId?: string;
  instagramUsername?: string;
}

interface SessionEntry {
  storeId: string;
  expiresAt: number;
  metaTokens?: MetaTokens;
}

// In-memory session store. Resets on server restart / Vercel cold start.
const sessions = new Map<string, SessionEntry>();

export function createSession(storeId: string): string {
  const token = crypto.randomUUID();
  sessions.set(token, {
    storeId,
    expiresAt: Date.now() + SESSION_TTL_MS,
  });
  return token;
}

export function getStoreIdFromSession(token: string): string | null {
  const entry = sessions.get(token);
  if (!entry) return null;
  if (Date.now() > entry.expiresAt) {
    sessions.delete(token);
    return null;
  }
  return entry.storeId;
}

export function getMetaTokens(token: string): MetaTokens | null {
  // First check in-memory session
  const entry = sessions.get(token);
  if (entry && Date.now() <= entry.expiresAt && entry.metaTokens) {
    return entry.metaTokens;
  }
  return null;
}

/**
 * Get meta tokens from the cookie value directly (survives server restarts).
 */
export function getMetaTokensFromCookie(cookieValue: string | undefined): MetaTokens | null {
  if (!cookieValue) return null;
  try {
    return JSON.parse(Buffer.from(cookieValue, "base64").toString("utf-8"));
  } catch {
    return null;
  }
}

/**
 * Encode meta tokens for cookie storage.
 */
export function encodeMetaTokensForCookie(metaTokens: MetaTokens): string {
  return Buffer.from(JSON.stringify(metaTokens)).toString("base64");
}

export function setMetaTokens(token: string, metaTokens: Partial<MetaTokens>): void {
  const entry = sessions.get(token);
  if (!entry) return;
  entry.metaTokens = { ...entry.metaTokens, ...metaTokens };
}

export function deleteSession(token: string): void {
  sessions.delete(token);
}
