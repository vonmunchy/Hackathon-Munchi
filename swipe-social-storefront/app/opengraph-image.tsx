import { ImageResponse } from "next/og";

export const runtime = "edge";

export const alt = "SwiftStore — Your Social Storefront";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          background: "linear-gradient(135deg, #1e1b4b 0%, #4c1d95 50%, #7c3aed 100%)",
          fontFamily: "sans-serif",
        }}
      >
        {/* Logo mark */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 100,
            height: 100,
            borderRadius: 20,
            backgroundColor: "#7c3aed",
            border: "3px solid rgba(255,255,255,0.3)",
            marginBottom: 32,
          }}
        >
          <span
            style={{
              fontSize: 60,
              fontWeight: 800,
              color: "white",
              lineHeight: 1,
            }}
          >
            S
          </span>
        </div>

        {/* Brand name */}
        <div
          style={{
            display: "flex",
            fontSize: 72,
            fontWeight: 800,
            color: "white",
            letterSpacing: "-2px",
            marginBottom: 16,
          }}
        >
          SwiftStore
        </div>

        {/* Tagline */}
        <div
          style={{
            display: "flex",
            fontSize: 28,
            color: "rgba(255,255,255,0.7)",
            fontWeight: 400,
          }}
        >
          Your Social Storefront
        </div>
      </div>
    ),
    { ...size }
  );
}
