// @vitest-environment node
import { describe, expect, it, vi } from "vitest";
import { withHandler } from "@/lib/server/handler";

vi.mock("@/lib/server/logger", () => ({
  logger: { info: vi.fn(), warn: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/server/profiler", () => ({
  createProfiler: vi.fn(() => ({ method: "GET", path: "/test", startTime: 0 })),
  finishProfiler: vi.fn((_ctx: unknown, res: Response) => res),
}));

describe("withHandler", () => {
  it("returns the handler response on success", async () => {
    const handler = withHandler("GET", "/test", () => new Response("ok"));
    const response = await handler();
    expect(await response.text()).toBe("ok");
  });

  it("catches errors and returns error response", async () => {
    const handler = withHandler("GET", "/fail", () => {
      throw new Error("boom");
    });
    const response = await handler();
    expect(response.status).toBe(500);
  });
});
