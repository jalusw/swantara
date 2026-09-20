import { afterEach, describe, expect, it } from "vitest";
import {
  accessTokenKey,
  csrfKey,
  hasAccessToken,
  hasSessionHint,
  refreshTokenKey,
} from "@/lib/constants/cookies";

describe("cookie constants", () => {
  it("defines access_token key", () => {
    expect(accessTokenKey).toBe("access_token");
  });

  it("defines refresh_token key", () => {
    expect(refreshTokenKey).toBe("refresh_token");
  });
});

describe("hasAccessToken", () => {
  afterEach(() => {
    document.cookie = `${csrfKey}=; max-age=0`;
  });

  it("returns true when session cookie exists", () => {
    document.cookie = `${csrfKey}=abc123`;
    expect(hasAccessToken()).toBe(true);
  });

  it("returns false when session cookie is absent", () => {
    expect(hasAccessToken()).toBe(false);
  });

  it("ignores HttpOnly access_token cookie (not visible to JS)", () => {
    document.cookie = "access_token=abc123";
    expect(hasSessionHint()).toBe(false);
  });
});
