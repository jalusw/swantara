import { describe, expect, it } from "vitest";
import { extractSessionTokens } from "../bff";

describe("extractSessionTokens", () => {
  it("reads camelCase tokens", () => {
    expect(extractSessionTokens({ data: { accessToken: "a", refreshToken: "r" } })).toEqual({
      accessToken: "a",
      refreshToken: "r",
    });
  });

  it("reads snake_case tokens", () => {
    expect(extractSessionTokens({ data: { access_token: "a", refresh_token: "r" } })).toEqual({
      accessToken: "a",
      refreshToken: "r",
    });
  });

  it("returns null when tokens are incomplete", () => {
    expect(extractSessionTokens({ data: { accessToken: "a" } })).toBeNull();
    expect(extractSessionTokens({})).toBeNull();
    expect(extractSessionTokens(null)).toBeNull();
  });
});
