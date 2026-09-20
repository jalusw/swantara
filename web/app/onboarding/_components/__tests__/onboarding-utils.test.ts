import { describe, expect, it } from "vitest";
import { resolveStandardCode } from "../onboarding-utils";

describe("resolveStandardCode", () => {
  it("returns PSAK_EMKM for an Indonesian business", () => {
    expect(resolveStandardCode("ID")).toBe("PSAK_EMKM");
  });

  it("returns IFRS for businesses outside Indonesia", () => {
    expect(resolveStandardCode("US")).toBe("IFRS");
    expect(resolveStandardCode("SG")).toBe("IFRS");
    expect(resolveStandardCode("MY")).toBe("IFRS");
  });

  it("matches lowercase country codes", () => {
    expect(resolveStandardCode("id")).toBe("PSAK_EMKM");
  });
});
