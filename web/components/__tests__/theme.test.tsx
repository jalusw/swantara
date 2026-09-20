import { describe, expect, it } from "vitest";
import { APP_THEMES, isAppTheme } from "@/components/theme";

describe("theme", () => {
  it("exports APP_THEMES with light and dark", () => {
    expect(APP_THEMES).toEqual(["light", "dark"]);
  });

  it("isAppTheme returns true for valid themes", () => {
    expect(isAppTheme("light")).toBe(true);
    expect(isAppTheme("dark")).toBe(true);
  });

  it("isAppTheme returns false for invalid themes", () => {
    expect(isAppTheme("system")).toBe(false);
    expect(isAppTheme("")).toBe(false);
    expect(isAppTheme("auto")).toBe(false);
  });
});
