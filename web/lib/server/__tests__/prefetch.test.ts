import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/server/api-client", () => ({
  createServerSwantaraService: vi.fn().mockResolvedValue({
    me: {
      me: vi.fn().mockResolvedValue({ data: {} }),
      organizations: vi.fn().mockResolvedValue({ data: [] }),
      permissions: vi.fn().mockResolvedValue({ data: [] }),
    },
  }),
}));

describe("prefetch utilities", () => {
  it("exports prefetchMeData", async () => {
    const { prefetchMeData } = await import("../prefetch");
    expect(typeof prefetchMeData).toBe("function");
  });

  it("exports prefetchPermissionsData", async () => {
    const { prefetchPermissionsData } = await import("../prefetch");
    expect(typeof prefetchPermissionsData).toBe("function");
  });

  it("prefetchMeData returns dehydrated state", async () => {
    const { prefetchMeData } = await import("../prefetch");
    const state = await prefetchMeData();
    expect(state).toHaveProperty("queries");
  });
});
