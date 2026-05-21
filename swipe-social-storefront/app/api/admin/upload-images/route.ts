import { NextResponse } from "next/server";
import convexServer from "@/lib/convex-server";
import { api } from "@/convex/_generated/api";
import fs from "fs";
import path from "path";

/**
 * One-time admin endpoint: uploads local product images to Convex Storage
 * and updates the product records with public URLs.
 */
export async function POST() {
  try {
    // Get all products
    const storeSlug = "island-finds-mv";
    const data = await convexServer.query(api.products.listByStoreSlugWithVariants, { storeSlug });
    if (!data?.products) {
      return NextResponse.json({ error: "No products found" }, { status: 404 });
    }

    const results = [];

    for (const product of data.products) {
      const localPath = product.imageUrls[0];
      if (!localPath || localPath.startsWith("http")) {
        results.push({ name: product.name, status: "skipped", reason: "already public URL or no image" });
        continue;
      }

      // Read the local file
      const publicDir = path.join(process.cwd(), "public");
      const filePath = path.join(publicDir, localPath);

      if (!fs.existsSync(filePath)) {
        results.push({ name: product.name, status: "error", reason: `File not found: ${filePath}` });
        continue;
      }

      // Generate upload URL
      const uploadUrl = await convexServer.mutation(api.storage.generateUploadUrl, {});

      // Read file and upload
      const fileBuffer = fs.readFileSync(filePath);
      const contentType = localPath.endsWith(".png") ? "image/png" : "image/jpeg";

      const uploadRes = await fetch(uploadUrl, {
        method: "POST",
        headers: { "Content-Type": contentType },
        body: fileBuffer,
      });

      if (!uploadRes.ok) {
        results.push({ name: product.name, status: "error", reason: `Upload failed: ${uploadRes.status}` });
        continue;
      }

      const { storageId } = await uploadRes.json();

      // Get the public URL
      const publicUrl = await convexServer.query(api.storage.getUrl, { storageId });

      if (publicUrl) {
        // Update the product with the public URL
        await convexServer.mutation(api.storage.updateProductImages, {
          productId: product._id,
          imageUrls: [publicUrl],
        });
        results.push({ name: product.name, status: "uploaded", publicUrl });
      } else {
        results.push({ name: product.name, status: "error", reason: "No public URL returned" });
      }
    }

    return NextResponse.json({ results });
  } catch (err) {
    console.error("Upload error:", err);
    return NextResponse.json(
      { error: err instanceof Error ? err.message : "Upload failed" },
      { status: 500 },
    );
  }
}
