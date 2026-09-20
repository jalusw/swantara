import { describe, expect, it } from "vitest";

type CountLine = {
  id: number;
  productName: string;
  theoreticalQty: number;
  countedQty: number;
  diffQty: number;
};

function computeDiffPreview(lines: CountLine[]) {
  const totalDiff = lines.reduce((sum, line) => sum + line.diffQty, 0);
  const hasDiff = lines.some((line) => line.diffQty !== 0);
  const diffCount = lines.filter((line) => line.diffQty !== 0).length;
  return { totalDiff, hasDiff, diffCount };
}

describe("count diff preview", () => {
  it("detects no differences when all diffs are zero", () => {
    const lines: CountLine[] = [
      {
        id: 1,
        productName: "A",
        theoreticalQty: 10,
        countedQty: 10,
        diffQty: 0,
      },
      { id: 2, productName: "B", theoreticalQty: 5, countedQty: 5, diffQty: 0 },
    ];
    const result = computeDiffPreview(lines);
    expect(result.hasDiff).toBe(false);
    expect(result.totalDiff).toBe(0);
    expect(result.diffCount).toBe(0);
  });

  it("detects positive differences", () => {
    const lines: CountLine[] = [
      {
        id: 1,
        productName: "A",
        theoreticalQty: 10,
        countedQty: 12,
        diffQty: 2,
      },
    ];
    const result = computeDiffPreview(lines);
    expect(result.hasDiff).toBe(true);
    expect(result.totalDiff).toBe(2);
    expect(result.diffCount).toBe(1);
  });

  it("detects negative differences", () => {
    const lines: CountLine[] = [
      {
        id: 1,
        productName: "A",
        theoreticalQty: 10,
        countedQty: 8,
        diffQty: -2,
      },
    ];
    const result = computeDiffPreview(lines);
    expect(result.hasDiff).toBe(true);
    expect(result.totalDiff).toBe(-2);
    expect(result.diffCount).toBe(1);
  });

  it("computes net diff across mixed positive and negative", () => {
    const lines: CountLine[] = [
      {
        id: 1,
        productName: "A",
        theoreticalQty: 10,
        countedQty: 12,
        diffQty: 2,
      },
      {
        id: 2,
        productName: "B",
        theoreticalQty: 5,
        countedQty: 3,
        diffQty: -2,
      },
      {
        id: 3,
        productName: "C",
        theoreticalQty: 8,
        countedQty: 10,
        diffQty: 2,
      },
    ];
    const result = computeDiffPreview(lines);
    expect(result.totalDiff).toBe(2);
    expect(result.diffCount).toBe(3);
  });

  it("handles empty lines", () => {
    const result = computeDiffPreview([]);
    expect(result.hasDiff).toBe(false);
    expect(result.totalDiff).toBe(0);
    expect(result.diffCount).toBe(0);
  });
});
