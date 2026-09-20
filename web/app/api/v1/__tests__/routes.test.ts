import { HttpResponse, http } from "msw";
import type { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { GET as catchAllGet } from "@/app/api/v1/[...path]/route";
import { callRoute, type RequestOptions, type RouteHandlerResult, server } from "@/lib/tests";

type Envelope = {
  success: boolean;
  message: string;
  data: Record<string, unknown>;
};

type Route = (request: NextRequest) => Response | Promise<Response>;

async function call(
  route: Route,
  options: RequestOptions = {},
): Promise<RouteHandlerResult & { body: Envelope }> {
  return callRoute(route, options) as Promise<RouteHandlerResult & { body: Envelope }>;
}

describe("app/api/v1 catch-all proxy route", () => {
  it("forwards requests to the service and returns the camelCased envelope", async () => {
    server.use(
      http.get("*/api/v1/me", () =>
        HttpResponse.json({
          success: true,
          message: "Profile retrieved.",
          data: { id: 1, email: "demo@swantara.local" },
        }),
      ),
    );

    const result = await call(
      (request) => catchAllGet(request, { params: Promise.resolve({ path: ["me"] }) }),
      { url: "http://localhost/api/v1/me" },
    );

    expect(result.status).toBe(200);
    expect(result.body.data).toEqual({
      id: 1,
      email: "demo@swantara.local",
    });
  });

  it("returns a service unavailable envelope when the service is down", async () => {
    server.use(http.get("*/api/v1/me", () => HttpResponse.error()));

    const result = await call(
      (request) => catchAllGet(request, { params: Promise.resolve({ path: ["me"] }) }),
      { url: "http://localhost/api/v1/me" },
    );

    expect(result.status).toBe(503);
    expect(result.body.success).toBe(false);
  });
});
