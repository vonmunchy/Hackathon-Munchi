import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getMetaTokens, getMetaTokensFromCookie } from "@/lib/session";
import { SESSION_COOKIE_NAME } from "@/lib/constants";

const GRAPH_API_VERSION = "v25.0";
const GRAPH_API_BASE = `https://graph.facebook.com/${GRAPH_API_VERSION}`;

export async function POST(request: NextRequest) {
  try {
    const cookieStore = await cookies();
    const token = cookieStore.get(SESSION_COOKIE_NAME)?.value;
    const metaTokensCookie = cookieStore.get("meta_tokens")?.value;

    if (!token && !metaTokensCookie) {
      return NextResponse.json(
        { error: "Authentication required" },
        { status: 401 },
      );
    }

    // Try in-memory first, then fall back to cookie
    let metaTokens = token ? getMetaTokens(token) : null;
    if (!metaTokens) {
      metaTokens = getMetaTokensFromCookie(metaTokensCookie);
    }
    const body = await request.json();
    const { platform, message, link, imageUrl } = body as {
      platform: "facebook" | "instagram";
      message: string;
      link?: string;
      imageUrl?: string;
    };

    if (!platform || !message) {
      return NextResponse.json(
        { error: "platform and message are required" },
        { status: 400 },
      );
    }

    if (platform === "facebook") {
      if (!metaTokens?.facebookPageAccessToken || !metaTokens?.facebookPageId) {
        return NextResponse.json(
          { error: "Facebook not connected. Please login with Facebook first." },
          { status: 403 },
        );
      }

      let result;
      if (imageUrl) {
        // Photo post — use /photos endpoint with image URL and caption
        const photoParams = new URLSearchParams({
          url: imageUrl,
          message,
          access_token: metaTokens.facebookPageAccessToken,
        });
        const res = await fetch(
          `${GRAPH_API_BASE}/${metaTokens.facebookPageId}/photos`,
          { method: "POST", body: photoParams },
        );
        if (!res.ok) {
          const error = await res.json();
          throw new Error(error.error?.message || res.statusText);
        }
        result = await res.json();
        result.postUrl = `https://www.facebook.com/${result.post_id || result.id}`;
      } else {
        // Text/link post — use /feed endpoint
        const feedParams = new URLSearchParams({
          message,
          access_token: metaTokens.facebookPageAccessToken,
        });
        if (link) {
          feedParams.set("link", link);
        }
        const res = await fetch(
          `${GRAPH_API_BASE}/${metaTokens.facebookPageId}/feed`,
          { method: "POST", body: feedParams },
        );
        if (!res.ok) {
          const error = await res.json();
          throw new Error(error.error?.message || res.statusText);
        }
        result = await res.json();
        result.postUrl = `https://www.facebook.com/${result.id}`;
      }

      return NextResponse.json({
        success: true,
        postId: result.id,
        postUrl: result.postUrl,
      });
    }

    if (platform === "instagram") {
      if (!metaTokens?.instagramAccessToken || !metaTokens?.instagramUserId) {
        return NextResponse.json(
          { error: "Instagram not connected. Please login with Instagram first." },
          { status: 403 },
        );
      }

      if (!imageUrl) {
        return NextResponse.json(
          { error: "Instagram publishing requires an image URL" },
          { status: 400 },
        );
      }

      // Step 1: Create media container
      const containerParams = new URLSearchParams({
        image_url: imageUrl,
        caption: message,
        access_token: metaTokens.instagramAccessToken,
      });

      const containerRes = await fetch(
        `${GRAPH_API_BASE}/${metaTokens.instagramUserId}/media`,
        { method: "POST", body: containerParams },
      );

      if (!containerRes.ok) {
        const error = await containerRes.json();
        throw new Error(error.error?.message || "Container creation failed");
      }

      const container = await containerRes.json();

      // Step 2: Publish container
      const publishParams = new URLSearchParams({
        creation_id: container.id,
        access_token: metaTokens.instagramAccessToken,
      });

      const publishRes = await fetch(
        `${GRAPH_API_BASE}/${metaTokens.instagramUserId}/media_publish`,
        { method: "POST", body: publishParams },
      );

      if (!publishRes.ok) {
        const error = await publishRes.json();
        throw new Error(error.error?.message || "Instagram publish failed");
      }

      const publishResult = await publishRes.json();

      return NextResponse.json({
        success: true,
        postId: publishResult.id,
      });
    }

    return NextResponse.json(
      { error: "Unsupported platform" },
      { status: 400 },
    );
  } catch (err) {
    console.error("Meta publish error:", err);
    const message = err instanceof Error ? err.message : "Failed to publish";
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
