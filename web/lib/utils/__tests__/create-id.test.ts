import { afterEach, describe, expect, it, vi } from "vitest";
import { createId } from "../create-id";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("createId", () => {
  it("uses crypto.randomUUID when available", () => {
    vi.stubGlobal("crypto", { randomUUID: () => "uuid-1" });
    expect(createId()).toBe("uuid-1");
  });

  it("falls back to a random string without randomUUID", () => {
    vi.stubGlobal("crypto", {});
    const first = createId();
    const second = createId();
    expect(first).toBeTruthy();
    expect(typeof first).toBe("string");
    expect(first).not.toBe(second);
  });
});
