import { describe, expect, it } from "vitest";
import type { Tax } from "@/lib/services/swantara";
import { formatTaxAmount, taxTypeTone } from "../tax-utils";

function tax(overrides: Partial<Tax>): Tax {
  return {
    id: overrides.id ?? 1,
    createdAt: new Date(),
    updatedAt: new Date(),
    organizationId: 1,
    name: overrides.name ?? "Pajak",
    amount: overrides.amount ?? null,
    type: overrides.type ?? "percent",
    scope: overrides.scope ?? "sale",
    priceInclude: false,
    taxAccountId: null,
    refundTaxAccountId: null,
    active: true,
    ...overrides,
  } as Tax;
}

describe("formatTaxAmount", () => {
  it("formats percent tax with % suffix", () => {
    const t = tax({ type: "percent", amount: 11 });
    expect(formatTaxAmount(t)).toBe("11%");
  });

  it("defaults to 0% when percent amount is null", () => {
    const t = tax({ type: "percent", amount: null });
    expect(formatTaxAmount(t)).toBe("0%");
  });

  it("formats fixed tax as plain number", () => {
    const t = tax({ type: "fixed", amount: 5000 });
    expect(formatTaxAmount(t)).toBe("5000");
  });

  it("defaults to 0 when fixed amount is null", () => {
    const t = tax({ type: "fixed", amount: null });
    expect(formatTaxAmount(t)).toBe("0");
  });

  it("returns Group for group tax type", () => {
    const t = tax({ type: "group", amount: null });
    expect(formatTaxAmount(t)).toBe("Group");
  });
});

describe("taxTypeTone", () => {
  it("returns correct tone for each tax type", () => {
    expect(taxTypeTone("percent")).toBe("info");
    expect(taxTypeTone("fixed")).toBe("warning");
    expect(taxTypeTone("group")).toBe("neutral");
  });
});
