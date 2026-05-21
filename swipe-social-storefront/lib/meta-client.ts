/**
 * Meta Graph API client for publishing to Facebook Page.
 * Uses Page Access Token from env vars.
 */

const META_PAGE_ID = process.env.META_PAGE_ID || "";
const META_PAGE_ACCESS_TOKEN = process.env.META_PAGE_ACCESS_TOKEN || "";
const META_IG_ACCOUNT_ID = process.env.META_IG_ACCOUNT_ID || "";
const GRAPH_API_VERSION = "v25.0";
const GRAPH_API_BASE = `https://graph.facebook.com/${GRAPH_API_VERSION}`;

export function isMetaConfigured(): boolean {
  return !!META_PAGE_ID && !!META_PAGE_ACCESS_TOKEN;
}

export function isInstagramConfigured(): boolean {
  return isMetaConfigured() && !!META_IG_ACCOUNT_ID;
}

/**
 * Publish a text + link post to the Facebook Page.
 */
export async function publishToFacebook(
  message: string,
  link?: string,
): Promise<{ id: string; postUrl: string }> {
  const params = new URLSearchParams({
    message,
    access_token: META_PAGE_ACCESS_TOKEN,
  });

  if (link) {
    params.set("link", link);
  }

  const res = await fetch(`${GRAPH_API_BASE}/${META_PAGE_ID}/feed`, {
    method: "POST",
    body: params,
  });

  if (!res.ok) {
    const error = await res.json();
    throw new Error(
      `Facebook publish failed: ${error.error?.message || res.statusText}`,
    );
  }

  const data = await res.json();
  // data.id is like "pageId_postId"
  const postId = data.id;
  const postUrl = `https://www.facebook.com/${postId}`;

  return { id: postId, postUrl };
}

/**
 * Publish a photo + caption post to the Facebook Page.
 */
export async function publishPhotoToFacebook(
  imageUrl: string,
  caption: string,
): Promise<{ id: string; postUrl: string }> {
  const params = new URLSearchParams({
    url: imageUrl,
    message: caption,
    access_token: META_PAGE_ACCESS_TOKEN,
  });

  const res = await fetch(`${GRAPH_API_BASE}/${META_PAGE_ID}/photos`, {
    method: "POST",
    body: params,
  });

  if (!res.ok) {
    const error = await res.json();
    throw new Error(
      `Facebook photo publish failed: ${error.error?.message || res.statusText}`,
    );
  }

  const data = await res.json();
  const postUrl = `https://www.facebook.com/${data.post_id || data.id}`;

  return { id: data.id, postUrl };
}

/**
 * Publish to Instagram (two-step: create container, then publish).
 * Requires META_IG_ACCOUNT_ID and a publicly accessible image URL.
 */
export async function publishToInstagram(
  imageUrl: string,
  caption: string,
): Promise<{ id: string }> {
  if (!META_IG_ACCOUNT_ID) {
    throw new Error("Instagram Business Account ID not configured");
  }

  // Step 1: Create media container
  const containerParams = new URLSearchParams({
    image_url: imageUrl,
    caption,
    access_token: META_PAGE_ACCESS_TOKEN,
  });

  const containerRes = await fetch(
    `${GRAPH_API_BASE}/${META_IG_ACCOUNT_ID}/media`,
    { method: "POST", body: containerParams },
  );

  if (!containerRes.ok) {
    const error = await containerRes.json();
    throw new Error(
      `Instagram container creation failed: ${error.error?.message || containerRes.statusText}`,
    );
  }

  const container = await containerRes.json();

  // Step 2: Publish container
  const publishParams = new URLSearchParams({
    creation_id: container.id,
    access_token: META_PAGE_ACCESS_TOKEN,
  });

  const publishRes = await fetch(
    `${GRAPH_API_BASE}/${META_IG_ACCOUNT_ID}/media_publish`,
    { method: "POST", body: publishParams },
  );

  if (!publishRes.ok) {
    const error = await publishRes.json();
    throw new Error(
      `Instagram publish failed: ${error.error?.message || publishRes.statusText}`,
    );
  }

  const result = await publishRes.json();
  return { id: result.id };
}
