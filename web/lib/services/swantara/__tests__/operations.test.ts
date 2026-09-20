import { describe, expect, it } from "vitest";
import { Employees, Health, Kpis, Reports } from "../operations";

describe("operations service classes", () => {
  it("exports Health with expected methods", () => {
    const proto = Health.prototype;
    expect(typeof proto.check).toBe("function");
  });

  it("exports Reports with expected methods", () => {
    const proto = Reports.prototype;
    expect(typeof proto.trialBalance).toBe("function");
    expect(typeof proto.aging).toBe("function");
    expect(typeof proto.profitAndLoss).toBe("function");
    expect(typeof proto.balanceSheet).toBe("function");
    expect(typeof proto.cashFlow).toBe("function");
  });

  it("exports Kpis with expected methods", () => {
    const proto = Kpis.prototype;
    expect(typeof proto.sales).toBe("function");
    expect(typeof proto.finance).toBe("function");
    expect(typeof proto.inventory).toBe("function");
    expect(typeof proto.payroll).toBe("function");
  });

  it("exports Employees with expected methods", () => {
    const proto = Employees.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.delete).toBe("function");
  });
});
