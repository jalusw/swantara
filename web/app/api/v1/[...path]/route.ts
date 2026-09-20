import type { NextRequest } from "next/server";
import { withHandler } from "@/lib/server/handler";
import { proxyRequest } from "@/lib/server/proxy";
import { createBadRequestResponse } from "@/lib/server/response";

type ProxyContext = {
  params: Promise<{ path: string[] }>;
};

function hasInvalidSegments(path: string[]): boolean {
  return path.some((segment) => {
    if (segment.length === 0 || segment.length > 128) return true;
    if (/[%?#\\]/.test(segment)) return true;
    let decoded = segment;
    try {
      decoded = decodeURIComponent(segment);
    } catch {
      return true;
    }
    const lower = decoded.toLowerCase();
    return (
      decoded.includes("\0") ||
      decoded === "." ||
      decoded === ".." ||
      lower === "%2e" ||
      decoded.includes("/") ||
      decoded.includes("\\") ||
      decoded.includes("..")
    );
  });
}

export const GET = withHandler(
  "GET",
  "/api/v1/[...path]",
  async (request: NextRequest, context: ProxyContext) => {
    const { path } = await context.params;
    if (hasInvalidSegments(path)) {
      return createBadRequestResponse("Invalid path segment.");
    }
    return proxyRequest(request, path);
  },
);

export const POST = withHandler(
  "POST",
  "/api/v1/[...path]",
  async (request: NextRequest, context: ProxyContext) => {
    const { path } = await context.params;
    if (hasInvalidSegments(path)) {
      return createBadRequestResponse("Invalid path segment.");
    }
    return proxyRequest(request, path);
  },
);

export const PUT = withHandler(
  "PUT",
  "/api/v1/[...path]",
  async (request: NextRequest, context: ProxyContext) => {
    const { path } = await context.params;
    if (hasInvalidSegments(path)) {
      return createBadRequestResponse("Invalid path segment.");
    }
    return proxyRequest(request, path);
  },
);

export const PATCH = withHandler(
  "PATCH",
  "/api/v1/[...path]",
  async (request: NextRequest, context: ProxyContext) => {
    const { path } = await context.params;
    if (hasInvalidSegments(path)) {
      return createBadRequestResponse("Invalid path segment.");
    }
    return proxyRequest(request, path);
  },
);

export const DELETE = withHandler(
  "DELETE",
  "/api/v1/[...path]",
  async (request: NextRequest, context: ProxyContext) => {
    const { path } = await context.params;
    if (hasInvalidSegments(path)) {
      return createBadRequestResponse("Invalid path segment.");
    }
    return proxyRequest(request, path);
  },
);
