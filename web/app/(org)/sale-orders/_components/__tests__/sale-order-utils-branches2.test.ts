import { describe, expect, it } from "vitest";
import type { PriceRule, SaleOrderLine } from "@/lib/services/swantara";
import {
  lineAmount,
  orderRemainingToBill,
  orderRemainingToShip,
  remainingToBill,
  remainingToShip,
  resolvePricePreview,
  saleOrderStateTone,
} from "../sale-order-utils";

function rule(over: Partial<PriceRule>): PriceRule {
  return {
    id: over.id ?? 1,
    priceBookId: 1,
    appliesTo: over.appliesTo ?? "all",
    itemId: over.itemId ?? null,
    categoryId: over.categoryId ?? null,
    minQty: over.minQty ?? 1,
    computeType: over.computeType ?? "fixed",
    fixedPrice: over.fixedPrice ?? null,
    discountPct: over.discountPct ?? null,
    dateStart: over.dateStart ?? null,
    dateEnd: over.dateEnd ?? null,
    createdAt: new Date(),
    updatedAt: new Date(),
  } as unknown as PriceRule;
}

function line(over: Partial<SaleOrderLine>): SaleOrderLine {
  return {
    qtyOrdered: 10,
    qtyDelivered: 0,
    qtyInvoiced: 0,
    qtyReturns: 0,
    ...over,
  } as SaleOrderLine;
}

describe("orderRemainingToShip / orderRemainingToBill", () => {
  it("sums remaining across lines", () => {
    const lines = [line({ qtyOrdered: 10, qtyDelivered: 4 }), line({ qtyOrdered: 5 })];
    expect(orderRemainingToShip({ amountTotal: 0 }, lines)).toBe(11);
  });

  it("sums remaining to bill across lines", () => {
    const lines = [line({ qtyOrdered: 10, qtyInvoiced: 7 }), line({ qtyOrdered: 5 })];
    expect(orderRemainingToBill({ amountTotal: 0 }, lines)).toBe(8);
  });

  it("returns zero for empty lines", () => {
    expect(orderRemainingToShip({ amountTotal: 0 }, [])).toBe(0);
    expect(orderRemainingToBill({ amountTotal: 0 }, [])).toBe(0);
  });
});

describe("saleOrderStateTone", () => {
  it("maps every known state", () => {
    expect(saleOrderStateTone("draft")).toBe("neutral");
    expect(saleOrderStateTone("sent")).toBe("info");
    expect(saleOrderStateTone("confirmed")).toBe("warning");
    expect(saleOrderStateTone("done")).toBe("success");
    expect(saleOrderStateTone("cancelled")).toBe("danger");
  });

  it("falls back to neutral for unknown states", () => {
    expect(saleOrderStateTone("archived" as never)).toBe("neutral");
  });
});

describe("resolvePricePreview uncovered branches", () => {
  const base = 100;
  const date = new Date("2026-06-15");

  it("returns base price when both variant and item are null", () => {
    expect(resolvePricePreview(null, null, 7, 5, date, base, [rule({})]).price).toBe(100);
  });

  it("skips rules below minQty and rules outside the date window", () => {
    const rules = [
      rule({ id: 1, minQty: 99, fixedPrice: 10, computeType: "fixed" }),
      rule({
        id: 2,
        fixedPrice: 20,
        computeType: "fixed",
        dateStart: new Date("2026-07-01"),
      }),
      rule({
        id: 3,
        fixedPrice: 30,
        computeType: "fixed",
        dateEnd: new Date("2026-01-01"),
      }),
    ];
    const result = resolvePricePreview(null, 10, null, 1, date, base, rules);
    expect(result.price).toBe(100);
    expect(result.ruleId).toBeNull();
  });

  it("ignores date bounds when the reference date is null", () => {
    const rules = [
      rule({
        id: 4,
        fixedPrice: 60,
        computeType: "fixed",
        dateStart: new Date("2026-07-01"),
        dateEnd: new Date("2026-01-01"),
      }),
    ];
    const result = resolvePricePreview(null, 10, null, 1, null, base, rules);
    expect(result.price).toBe(60);
    expect(result.ruleId).toBe(4);
  });

  it("filters variant, item, and category scope mismatches", () => {
    const rules = [
      rule({ id: 1, appliesTo: "variant", itemId: 999, fixedPrice: 10 }),
      rule({ id: 2, appliesTo: "item", itemId: 999, fixedPrice: 20 }),
      rule({ id: 3, appliesTo: "category", categoryId: 999, fixedPrice: 30 }),
    ];
    expect(resolvePricePreview(5, 10, 7, 1, date, base, rules).price).toBe(100);
  });

  it("matches a category rule when the category agrees", () => {
    const rules = [rule({ id: 9, appliesTo: "category", categoryId: 7, fixedPrice: 77 })];
    const result = resolvePricePreview(null, 10, 7, 1, date, base, rules);
    expect(result.price).toBe(77);
    expect(result.ruleId).toBe(9);
  });

  it("returns base price with the rule id when compute type is unknown", () => {
    const rules = [rule({ id: 5, computeType: "other" as never })];
    const result = resolvePricePreview(null, 10, null, 1, date, base, rules);
    expect(result.price).toBe(100);
    expect(result.ruleId).toBe(5);
  });

  it("returns base price when fixed rule has no fixed price", () => {
    const rules = [rule({ id: 6, computeType: "fixed", fixedPrice: null })];
    const result = resolvePricePreview(null, 10, null, 1, date, base, rules);
    expect(result.price).toBe(100);
    expect(result.ruleId).toBe(6);
  });

  it("returns base price when percent rule has no discount", () => {
    const rules = [rule({ id: 7, computeType: "percent", discountPct: null })];
    const result = resolvePricePreview(null, 10, null, 1, date, base, rules);
    expect(result.price).toBe(100);
    expect(result.ruleId).toBe(7);
  });

  it("handles missing qtyReturns on lines", () => {
    expect(remainingToShip({ qtyOrdered: 10, qtyDelivered: 3 } as SaleOrderLine)).toBe(7);
    expect(remainingToBill({ qtyOrdered: 10, qtyInvoiced: 3 } as SaleOrderLine)).toBe(7);
  });

  it("rounds line amounts for fractional discounts", () => {
    expect(lineAmount(3, 19.99, 33)).toBe(40.18);
    expect(lineAmount(0, 100, 10)).toBe(0);
  });
});
