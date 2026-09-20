import { describe, expect, it } from "vitest";
import { bomLineTotalQty } from "../products-data";

describe("bomLineTotalQty", () => {
  it("returns the same qty when scrap is zero", () => {
    expect(bomLineTotalQty({ qty: 10, scrapPct: 0 })).toBe(10);
  });

  it("adds 10% scrap to the qty", () => {
    expect(bomLineTotalQty({ qty: 10, scrapPct: 10 })).toBeCloseTo(11);
  });

  it("adds 2.5% scrap to the qty", () => {
    expect(bomLineTotalQty({ qty: 4, scrapPct: 2.5 })).toBeCloseTo(4.1);
  });

  it("doubles the qty at 100% scrap", () => {
    expect(bomLineTotalQty({ qty: 5, scrapPct: 100 })).toBe(10);
  });

  it("handles zero qty with scrap", () => {
    expect(bomLineTotalQty({ qty: 0, scrapPct: 10 })).toBe(0);
  });

  it("handles fractional qty with scrap", () => {
    expect(bomLineTotalQty({ qty: 2.5, scrapPct: 20 })).toBeCloseTo(3);
  });
});
