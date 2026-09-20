import { describe, expect, it } from "vitest";
import { buildLoginPayload } from "../factories";

describe("buildLoginPayload", () => {
  it("returns the default demo credentials", () => {
    expect(buildLoginPayload()).toEqual({ email: "demo@swantara.local", password: "secret" });
  });

  it("applies overrides", () => {
    expect(buildLoginPayload({ email: "user@example.com" })).toMatchObject({
      email: "user@example.com",
      password: "secret",
    });
  });
});
