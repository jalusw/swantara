import { describe, expect, it } from "vitest";
import { Invoices, JournalEntries, Payments, TaxPeriods } from "../finance";

describe("finance service classes", () => {
  it("exports JournalEntries with expected methods", () => {
    const proto = JournalEntries.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.reverse).toBe("function");
  });

  it("exports Invoices with expected methods", () => {
    const proto = Invoices.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.creditNote).toBe("function");
  });

  it("exports Payments with expected methods", () => {
    const proto = Payments.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.createOutbound).toBe("function");
  });

  it("exports TaxPeriods with expected methods", () => {
    const proto = TaxPeriods.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.close).toBe("function");
    expect(typeof proto.lock).toBe("function");
    expect(typeof proto.open).toBe("function");
  });
});
