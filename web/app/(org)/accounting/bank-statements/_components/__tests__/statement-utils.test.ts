import { describe, expect, it } from "vitest";
import { statementBalance } from "../statement-utils";

describe("statementBalance", () => {
  it("computes balance from start balance plus line amounts", () => {
    const result = statementBalance(1000, [{ amount: 200 }, { amount: -50 }, { amount: 300 }]);

    expect(result).toBe(1450);
  });

  it("returns start balance when no lines", () => {
    expect(statementBalance(500, [])).toBe(500);
  });

  it("handles negative start balance", () => {
    const result = statementBalance(-100, [{ amount: 250 }]);

    expect(result).toBe(150);
  });

  it("handles zero start balance", () => {
    const result = statementBalance(0, [{ amount: 100 }, { amount: -100 }]);

    expect(result).toBe(0);
  });
});
