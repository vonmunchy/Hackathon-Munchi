import { ConvexHttpClient } from "convex/browser";

// Singleton server-side Convex client for use in Next.js API routes.
// API routes MUST use this — useQuery/useMutation are React client-side only.
const convexServer = new ConvexHttpClient(process.env.NEXT_PUBLIC_CONVEX_URL!);

export default convexServer;
