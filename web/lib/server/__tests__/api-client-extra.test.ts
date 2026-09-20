import { describe, expect, it, vi } from "vitest";
import { createServerSwantaraService } from "../api-client";
import { serverConfig } from "../config";

vi.mock("next/headers", () => ({
  cookies: vi.fn().mockResolvedValue({
    get: (name: string) => (name === "access_token" ? { value: "extra-token" } : undefined),
  }),
}));

describe("createServerSwantaraService extra", () => {
  it("targets the configured service origin", async () => {
    const service = await createServerSwantaraService();
    expect(service).toBeDefined();
    expect(new URL(serverConfig.serviceApiUrl).origin).toContain("http");
  });

  it("exposes organization modules", async () => {
    const service = await createServerSwantaraService();
    expect(service.organizations).toBeDefined();
    expect(typeof service.me.organizations).toBe("function");
    expect(typeof service.me.permissions).toBe("function");
  });
});
