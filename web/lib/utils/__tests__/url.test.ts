import { describe, expect, it } from "vitest";
import { isProtectedPage } from "../url";

describe("isProtectedPage", () => {
  it("protects app pages", () => {
    expect(isProtectedPage("/onboarding")).toBe(true);
    expect(isProtectedPage("/dashboard")).toBe(true);
    expect(isProtectedPage("/crm/opportunities")).toBe(true);
  });

  it("leaves public pages alone", () => {
    expect(isProtectedPage("/login")).toBe(false);
    expect(isProtectedPage("/register")).toBe(false);
    expect(isProtectedPage("/")).toBe(false);
  });
});
