import { describe, expect, it, vi } from "vitest";

const meHandlers = {
  me: {
    me: vi.fn().mockResolvedValue({ data: {} }),
    organizations: vi.fn().mockResolvedValue({ data: [] }),
    permissions: vi.fn().mockResolvedValue({ data: [] }),
  },
};

vi.mock("@/lib/server/api-client", () => ({
  createServerSwantaraService: vi.fn().mockResolvedValue(meHandlers),
}));

describe("prefetch utilities extra", () => {
  it("prefetchPermissionsData returns dehydrated state for an organization", async () => {
    const { prefetchPermissionsData } = await import("../prefetch");
    const state = await prefetchPermissionsData(1);
    expect(state.queries.length).toBeGreaterThan(0);
    expect(meHandlers.me.permissions).toHaveBeenCalledWith(1);
  });

  it("prefetchMeData tolerates a failing organizations query", async () => {
    meHandlers.me.organizations.mockRejectedValueOnce(new Error("down"));
    const { prefetchMeData } = await import("../prefetch");
    const state = await prefetchMeData();
    expect(state).toHaveProperty("queries");
    expect(state.queries.length).toBeGreaterThanOrEqual(1);
  });
});
