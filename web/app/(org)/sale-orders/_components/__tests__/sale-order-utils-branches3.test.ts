import { describe, expect, it } from "vitest";
import type { PriceRule } from "@/lib/services/swantara";
import {
  canCancel,
  canConfirm,
  canDeliver,
  canDone,
  canEdit,
  canInvoice,
  canSend,
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

const baseDate = new Date("2026-06-15");

describe("sale-order-utils branches3", () => {
  it("treats missing returns as zero", () => {
    expect(remainingToShip({ qtyOrdered: 8, qtyDelivered: 3 } as never)).toBe(5);
    expect(remainingToBill({ qtyOrdered: 8, qtyInvoiced: 3 } as never)).toBe(5);
  });

  it("sums remaining across lines", () => {
    const lines = [
      { qtyOrdered: 5, qtyDelivered: 2, qtyReturns: 0, qtyInvoiced: 1 },
      { qtyOrdered: 4, qtyDelivered: 4, qtyReturns: 0, qtyInvoiced: 4 },
    ] as never[];
    expect(orderRemainingToShip({ amountTotal: 0 }, lines as never)).toBe(3);
    expect(orderRemainingToBill({ amountTotal: 0 }, lines as never)).toBe(4);
    expect(orderRemainingToShip({ amountTotal: 0 }, [])).toBe(0);
  });

  it("covers every state gate negative branch", () => {
    expect(canSend("cancelled")).toBe(false);
    expect(canConfirm("cancelled")).toBe(false);
    expect(canCancel("cancelled")).toBe(false);
    expect(canDone("cancelled")).toBe(false);
    expect(canEdit("cancelled")).toBe(false);
    expect(canDeliver("cancelled")).toBe(false);
    expect(canInvoice("cancelled")).toBe(false);
    expect(canInvoice("sent")).toBe(false);
  });

  it("maps every state tone including fallback", () => {
    expect(saleOrderStateTone("draft")).toBe("neutral");
    expect(saleOrderStateTone("sent")).toBe("info");
    expect(saleOrderStateTone("confirmed")).toBe("warning");
    expect(saleOrderStateTone("done")).toBe("success");
    expect(saleOrderStateTone("cancelled")).toBe("danger");
    expect(saleOrderStateTone("archived" as never)).toBe("neutral");
  });

  it("returns base price when both ids are null", () => {
    const out = resolvePricePreview(null, null, 7, 2, baseDate, 100, [
      rule({ id: 9, fixedPrice: 10 }),
    ]);
    expect(out).toEqual({ price: 100, ruleId: null });
  });

  it("filters variant mismatch", () => {
    const out = resolvePricePreview(1, 10, null, 1, baseDate, 100, [
      rule({ id: 1, appliesTo: "variant", itemId: 2, fixedPrice: 10 }),
    ]);
    expect(out.price).toBe(100);
  });

  it("filters item mismatch", () => {
    const out = resolvePricePreview(null, 10, null, 1, baseDate, 100, [
      rule({ id: 1, appliesTo: "item", itemId: 99, fixedPrice: 10 }),
    ]);
    expect(out.price).toBe(100);
  });

  it("filters category mismatch", () => {
    const out = resolvePricePreview(null, 10, 5, 1, baseDate, 100, [
      rule({ id: 1, appliesTo: "category", categoryId: 6, fixedPrice: 10 }),
    ]);
    expect(out.price).toBe(100);
  });

  it("ignores date bounds when date is null", () => {
    const out = resolvePricePreview(null, 10, null, 1, null, 100, [
      rule({
        id: 1,
        appliesTo: "all",
        fixedPrice: 60,
        dateStart: new Date("2030-01-01"),
        dateEnd: new Date("2000-01-01"),
      }),
    ]);
    expect(out.price).toBe(60);
  });

  it("filters rules starting after the date", () => {
    const out = resolvePricePreview(null, 10, null, 1, new Date("2026-01-01"), 100, [
      rule({ id: 1, fixedPrice: 60, dateStart: new Date("2026-06-01") }),
    ]);
    expect(out.price).toBe(100);
  });

  it("filters rules ending before the date", () => {
    const out = resolvePricePreview(null, 10, null, 1, new Date("2026-06-15"), 100, [
      rule({ id: 1, fixedPrice: 60, dateEnd: new Date("2026-01-01") }),
    ]);
    expect(out.price).toBe(100);
  });

  it("falls back to base price for unknown compute type", () => {
    const out = resolvePricePreview(null, 10, null, 1, baseDate, 100, [
      rule({ id: 4, computeType: "formula" as never }),
    ]);
    expect(out).toEqual({ price: 100, ruleId: 4 });
  });

  it("falls back when fixed price is null and percent is null", () => {
    const fixedNull = resolvePricePreview(null, 10, null, 1, baseDate, 100, [
      rule({ id: 5, computeType: "fixed", fixedPrice: null }),
    ]);
    expect(fixedNull).toEqual({ price: 100, ruleId: 5 });
    const pctNull = resolvePricePreview(null, 10, null, 1, baseDate, 100, [
      rule({ id: 6, computeType: "percent", discountPct: null }),
    ]);
    expect(pctNull).toEqual({ price: 100, ruleId: 6 });
  });

  it("rounds percent and line amounts", () => {
    expect(
      resolvePricePreview(null, 10, null, 1, baseDate, 100, [
        rule({ id: 7, computeType: "percent", discountPct: 12.5 }),
      ]).price,
    ).toBe(87.5);
    expect(lineAmount(3, 19.99, 33)).toBe(40.18);
  });
});
