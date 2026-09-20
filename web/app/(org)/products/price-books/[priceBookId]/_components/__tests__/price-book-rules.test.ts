import { describe, expect, it } from "vitest";
import type { PriceRule } from "@/lib/services/swantara";

const scopePrecedence: Record<PriceRule["appliesTo"], number> = {
  variant: 0,
  item: 1,
  category: 2,
  all: 3,
};

function sortByPrecedence(rules: PriceRule[]): PriceRule[] {
  return [...rules].sort((a, b) => scopePrecedence[a.appliesTo] - scopePrecedence[b.appliesTo]);
}

describe("price_book rule precedence", () => {
  it("sorts rules from most specific to least specific", () => {
    const rules: PriceRule[] = [
      {
        id: 1,
        priceBookId: 10,
        appliesTo: "all",
        itemId: null,
        categoryId: null,
        minQty: 1,
        computeType: "fixed",
        fixedPrice: 10,
        discountPct: null,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 2,
        priceBookId: 10,
        appliesTo: "variant",
        itemId: 100,
        categoryId: null,
        minQty: 1,
        computeType: "fixed",
        fixedPrice: 8,
        discountPct: null,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 3,
        priceBookId: 10,
        appliesTo: "category",
        itemId: null,
        categoryId: 200,
        minQty: 1,
        computeType: "percent",
        fixedPrice: null,
        discountPct: 10,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 4,
        priceBookId: 10,
        appliesTo: "item",
        itemId: 100,
        categoryId: null,
        minQty: 5,
        computeType: "fixed",
        fixedPrice: 9,
        discountPct: null,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ];

    const sorted = sortByPrecedence(rules);
    expect(sorted.map((r) => r.appliesTo)).toEqual(["variant", "item", "category", "all"]);
  });

  it("preserves original order for rules with the same scope", () => {
    const rules: PriceRule[] = [
      {
        id: 1,
        priceBookId: 10,
        appliesTo: "item",
        itemId: 100,
        categoryId: null,
        minQty: 1,
        computeType: "fixed",
        fixedPrice: 10,
        discountPct: null,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 2,
        priceBookId: 10,
        appliesTo: "item",
        itemId: 200,
        categoryId: null,
        minQty: 1,
        computeType: "fixed",
        fixedPrice: 12,
        discountPct: null,
        dateStart: null,
        dateEnd: null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ];
    const sorted = sortByPrecedence(rules);
    expect(sorted[0]!.id).toBe(1);
    expect(sorted[1]!.id).toBe(2);
  });
});

describe("price_book rule date windows", () => {
  it("flags overlapping date windows on the same scope", () => {
    const rules: PriceRule[] = [
      {
        id: 1,
        priceBookId: 10,
        appliesTo: "item",
        itemId: 100,
        categoryId: null,
        minQty: 1,
        computeType: "fixed",
        fixedPrice: 10,
        discountPct: null,
        dateStart: new Date("2026-01-01"),
        dateEnd: new Date("2026-06-30"),
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 2,
        priceBookId: 10,
        appliesTo: "item",
        itemId: 100,
        categoryId: null,
        minQty: 1,
        computeType: "percent",
        fixedPrice: null,
        discountPct: 5,
        dateStart: new Date("2026-03-01"),
        dateEnd: new Date("2026-09-30"),
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ];

    function hasOverlap(ruleA: PriceRule, ruleB: PriceRule): boolean {
      if (ruleA.appliesTo !== ruleB.appliesTo) return false;
      if (ruleA.appliesTo === "item" && ruleA.itemId !== ruleB.itemId) return false;
      if (ruleA.appliesTo === "category" && ruleA.categoryId !== ruleB.categoryId) return false;
      const aStart = ruleA.dateStart ? new Date(ruleA.dateStart).getTime() : 0;
      const aEnd = ruleA.dateEnd ? new Date(ruleA.dateEnd).getTime() : Number.MAX_SAFE_INTEGER;
      const bStart = ruleB.dateStart ? new Date(ruleB.dateStart).getTime() : 0;
      const bEnd = ruleB.dateEnd ? new Date(ruleB.dateEnd).getTime() : Number.MAX_SAFE_INTEGER;
      return aStart <= bEnd && bStart <= aEnd;
    }

    expect(hasOverlap(rules[0]!, rules[1]!)).toBe(true);
  });

  it("does not flag non-overlapping date windows", () => {
    const ruleA: PriceRule = {
      id: 1,
      priceBookId: 10,
      appliesTo: "item",
      itemId: 100,
      categoryId: null,
      minQty: 1,
      computeType: "fixed",
      fixedPrice: 10,
      discountPct: null,
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-03-31"),
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    const ruleB: PriceRule = {
      id: 2,
      priceBookId: 10,
      appliesTo: "item",
      itemId: 100,
      categoryId: null,
      minQty: 1,
      computeType: "fixed",
      fixedPrice: 12,
      discountPct: null,
      dateStart: new Date("2026-04-01"),
      dateEnd: new Date("2026-06-30"),
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    function hasOverlap(ruleA: PriceRule, ruleB: PriceRule): boolean {
      if (ruleA.appliesTo !== ruleB.appliesTo) return false;
      if (ruleA.appliesTo === "item" && ruleA.itemId !== ruleB.itemId) return false;
      const aStart = ruleA.dateStart ? new Date(ruleA.dateStart).getTime() : 0;
      const aEnd = ruleA.dateEnd ? new Date(ruleA.dateEnd).getTime() : Number.MAX_SAFE_INTEGER;
      const bStart = ruleB.dateStart ? new Date(ruleB.dateStart).getTime() : 0;
      const bEnd = ruleB.dateEnd ? new Date(ruleB.dateEnd).getTime() : Number.MAX_SAFE_INTEGER;
      return aStart <= bEnd && bStart <= aEnd;
    }

    expect(hasOverlap(ruleA, ruleB)).toBe(false);
  });
});
