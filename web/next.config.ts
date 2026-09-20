import type { NextConfig } from "next";

const allowedDevOrigins = process.env?.ALLOWED_DEV_ORIGINS
  ? process.env?.ALLOWED_DEV_ORIGINS?.split(",").map((origin) => origin)
  : [];

let nextConfig: NextConfig = {
  output: "standalone" as const,
  allowedDevOrigins: [...allowedDevOrigins],
  logging: {
    fetches: {
      fullUrl: false,
    },
    incomingRequests: {
      ignore: [/^\/api\//],
    },
  },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Strict-Transport-Security",
            value: "max-age=63072000; includeSubDomains; preload",
          },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
        ],
      },
    ];
  },
  async redirects() {
    return [
      {
        source: "/org/:id",
        destination: "/org/:id/dashboard",
        permanent: true,
      },
    ];
  },
};

if (process.env.ANALYZE === "true") {
  const withBundleAnalyzer = require("@next/bundle-analyzer");
  nextConfig = withBundleAnalyzer({ enabled: true })(nextConfig);
}

export default nextConfig;
