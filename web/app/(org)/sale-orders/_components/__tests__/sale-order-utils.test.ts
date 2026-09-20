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
  remainingToBill,
  remainingToShip,
  resolvePricePreview,
} from "../sale-order-utils";

describe("remainingToShip / remainingToBill", () => {
  it("derives remaining to ship from ordered/delivered", () => {
    expect(remainingToShip({ qtyOrdered: 10, qtyDelivered: 4, qtyReturns: 0 })).toBe(6);
  });

  it("derives remaining to bill from ordered/invoiced", () => {
    expect(remainingToBill({ qtyOrdered: 10, qtyInvoiced: 7, qtyReturns: 0 })).toBe(3);
  });

  it("accounts for returns in net calculation", () => {
    expect(remainingToShip({ qtyOrdered: 10, qtyDelivered: 8, qtyReturns: 2 })).toBe(4);
    expect(remainingToBill({ qtyOrdered: 10, qtyInvoiced: 8, qtyReturns: 2 })).toBe(4);
  });

  it("never returns negative", () => {
    expect(remainingToShip({ qtyOrdered: 5, qtyDelivered: 10, qtyReturns: 0 })).toBe(0);
    expect(remainingToBill({ qtyOrdered: 5, qtyInvoiced: 10, qtyReturns: 0 })).toBe(0);
  });

  it("shows zero when fully delivered/invoiced", () => {
    expect(remainingToShip({ qtyOrdered: 5, qtyDelivered: 5, qtyReturns: 0 })).toBe(0);
    expect(remainingToBill({ qtyOrdered: 5, qtyInvoiced: 5, qtyReturns: 0 })).toBe(0);
  });
});

describe("sale order state gating", () => {
  it("allows send only in draft", () => {
    expect(canSend("draft")).toBe(true);
    expect(canSend("sent")).toBe(false);
    expect(canSend("confirmed")).toBe(false);
    expect(canSend("done")).toBe(false);
    expect(canSend("cancelled")).toBe(false);
  });

  it("allows confirm in draft or sent", () => {
    expect(canConfirm("draft")).toBe(true);
    expect(canConfirm("sent")).toBe(true);
    expect(canConfirm("confirmed")).toBe(false);
    expect(canConfirm("done")).toBe(false);
  });

  it("allows cancel in draft/sent/confirmed", () => {
    expect(canCancel("draft")).toBe(true);
    expect(canCancel("sent")).toBe(true);
    expect(canCancel("confirmed")).toBe(true);
    expect(canCancel("done")).toBe(false);
    expect(canCancel("cancelled")).toBe(false);
  });

  it("allows done only in confirmed", () => {
    expect(canDone("confirmed")).toBe(true);
    expect(canDone("draft")).toBe(false);
    expect(canDone("sent")).toBe(false);
    expect(canDone("done")).toBe(false);
  });

  it("allows edit only in draft", () => {
    expect(canEdit("draft")).toBe(true);
    expect(canEdit("sent")).toBe(false);
    expect(canEdit("confirmed")).toBe(false);
  });

  it("allows deliver only in confirmed", () => {
    expect(canDeliver("confirmed")).toBe(true);
    expect(canDeliver("draft")).toBe(false);
    expect(canDeliver("done")).toBe(false);
  });

  it("allows invoice in confirmed or done", () => {
    expect(canInvoice("confirmed")).toBe(true);
    expect(canInvoice("done")).toBe(true);
    expect(canInvoice("draft")).toBe(false);
    expect(canInvoice("sent")).toBe(false);
  });
});

describe("resolvePricePreview honors price_book", () => {
  const base = 100;
  const date = new Date("2026-06-15");

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

  it("returns base price when no rules match", () => {
    const r = resolvePricePreview(null, 10, null, 1, date, base, []);
    expect(r.price).toBe(100);
    expect(r.ruleId).toBeNull();
  });

  it("applies fixed price for most specific scope", () => {
    const rules = [
      rule({ id: 1, appliesTo: "all", fixedPrice: 90, computeType: "fixed" }),
      rule({
        id: 2,
        appliesTo: "item",
        itemId: 10,
        fixedPrice: 80,
        computeType: "fixed",
      }),
      rule({
        id: 3,
        appliesTo: "variant",
        itemId: 5,
        fixedPrice: 70,
        computeType: "fixed",
      }),
    ];
    const r = resolvePricePreview(5, 10, null, 1, date, base, rules);
    expect(r.price).toBe(70);
    expect(r.ruleId).toBe(3);
  });

  it("applies percent discount", () => {
    const rules = [
      rule({
        id: 2,
        appliesTo: "item",
        itemId: 10,
        discountPct: 10,
        computeType: "percent",
      }),
    ];
    const r = resolvePricePreview(null, 10, null, 1, date, base, rules);
    expect(r.price).toBe(90);
  });

  it("respects minQty", () => {
    const rules = [
      rule({
        id: 1,
        appliesTo: "all",
        minQty: 10,
        fixedPrice: 50,
        computeType: "fixed",
      }),
    ];
    expect(resolvePricePreview(null, 10, null, 5, date, base, rules).price).toBe(100);
    expect(resolvePricePreview(null, 10, null, 10, date, base, rules).price).toBe(50);
  });

  it("respects date window", () => {
    const rules = [
      rule({
        id: 1,
        appliesTo: "all",
        fixedPrice: 50,
        computeType: "fixed",
        dateStart: new Date("2026-01-01"),
        dateEnd: new Date("2026-01-31"),
      }),
    ];
    expect(resolvePricePreview(null, 10, null, 1, new Date("2026-02-01"), base, rules).price).toBe(
      100,
    );
    expect(resolvePricePreview(null, 10, null, 1, new Date("2026-01-15"), base, rules).price).toBe(
      50,
    );
  });

  it("prefers specific scope over generic regardless of order", () => {
    const rules = [
      rule({ id: 1, appliesTo: "all", fixedPrice: 95, computeType: "fixed" }),
      rule({
        id: 2,
        appliesTo: "category",
        categoryId: 5,
        fixedPrice: 85,
        computeType: "fixed",
      }),
    ];
    const r = resolvePricePreview(null, 10, 5, 1, date, base, rules);
    expect(r.price).toBe(85);
  });
});

describe("lineAmount", () => {
  it("computes qty * unitPrice with discount", () => {
    expect(lineAmount(10, 100, 0)).toBe(1000);
    expect(lineAmount(10, 100, 10)).toBe(900);
    expect(lineAmount(2, 50, 50)).toBe(50);
  });
});
