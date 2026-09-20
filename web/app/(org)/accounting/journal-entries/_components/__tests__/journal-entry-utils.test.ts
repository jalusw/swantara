import { describe, expect, it } from "vitest";
import { canPost, canReverse, moveBalance, moveStateTone } from "../journal-entry-utils";

describe("moveBalance", () => {
  it("detects balanced lines", () => {
    const result = moveBalance([
      { debit: 100, credit: 0 },
      { debit: 0, credit: 100 },
    ]);

    expect(result.totalDebit).toBe(100);
    expect(result.totalCredit).toBe(100);
    expect(result.isBalanced).toBe(true);
  });

  it("detects unbalanced lines", () => {
    const result = moveBalance([
      { debit: 100, credit: 0 },
      { debit: 0, credit: 50 },
    ]);

    expect(result.totalDebit).toBe(100);
    expect(result.totalCredit).toBe(50);
    expect(result.isBalanced).toBe(false);
  });

  it("handles empty lines", () => {
    const result = moveBalance([]);

    expect(result.totalDebit).toBe(0);
    expect(result.totalCredit).toBe(0);
    expect(result.isBalanced).toBe(true);
  });

  it("handles single line", () => {
    const result = moveBalance([{ debit: 50, credit: 50 }]);

    expect(result.totalDebit).toBe(50);
    expect(result.totalCredit).toBe(50);
    expect(result.isBalanced).toBe(true);
  });

  it("treats tiny difference as balanced", () => {
    const result = moveBalance([
      { debit: 100, credit: 0 },
      { debit: 0, credit: 100.005 },
    ]);

    expect(result.isBalanced).toBe(true);
  });
});

describe("canPost", () => {
  it("allows posting only in draft", () => {
    expect(canPost("draft")).toBe(true);
    expect(canPost("posted")).toBe(false);
    expect(canPost("cancelled")).toBe(false);
  });
});

describe("canReverse", () => {
  it("allows reversal only in posted", () => {
    expect(canReverse("posted")).toBe(true);
    expect(canReverse("draft")).toBe(false);
    expect(canReverse("cancelled")).toBe(false);
  });
});

describe("moveStateTone", () => {
  it("returns correct tone for each state", () => {
    expect(moveStateTone("draft")).toBe("neutral");
    expect(moveStateTone("posted")).toBe("success");
    expect(moveStateTone("cancelled")).toBe("danger");
  });
});
