import { describe, expect, it, vi } from "vitest";
import { createServerSwantaraService } from "../api-client";

vi.mock("next/headers", () => ({
  cookies: vi.fn().mockResolvedValue({
    get: (name: string) => (name === "access_token" ? { value: "tok" } : undefined),
  }),
}));

describe("createServerSwantaraService", () => {
  it("returns a SwantaraService instance", async () => {
    const service = await createServerSwantaraService();
    expect(service).toBeDefined();
  });

  it("exposes the me resource module", async () => {
    const service = await createServerSwantaraService();
    expect(service.me).toBeDefined();
  });
});
