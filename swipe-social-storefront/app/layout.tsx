import type { Metadata } from "next";
import { Inter, DM_Sans, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { ConvexClientProvider } from "./providers";
import { LayoutRouter } from "@/components/shells/layout-router";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700", "800"],
  display: "swap",
});

const dmSans = DM_Sans({
  variable: "--font-dm-sans",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  display: "swap",
});

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  weight: ["400", "500"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "SwiftStore",
  description:
    "Shop, sell, and trade crypto on one platform. Social commerce and P2P exchange powered by Swipe, built for the Maldives.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${dmSans.variable} ${jetBrainsMono.variable} h-full`}
    >
      <body className="min-h-full flex flex-col antialiased">
        <ConvexClientProvider>
          <LayoutRouter>{children}</LayoutRouter>
        </ConvexClientProvider>
      </body>
    </html>
  );
}
