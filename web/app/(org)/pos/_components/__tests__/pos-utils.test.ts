import { describe, expect, it } from "vitest";
import {
  canCloseSession,
  canInvoiceOrder,
  canPlaceOrder,
  canRefundOrder,
  canStartClosing,
  configDisplayName,
  orderLineSubtotal,
  orderTaxTotal,
  orderTotal,
  posOrderStateTone,
  posSessionStateTone,
  sessionOrderTotal,
  sessionPaymentSummary,
} from "../pos-utils";

describe("posSessionStateTone", () => {
  it("returns success for opened", () => {
    expect(posSessionStateTone("opened")).toBe("success");
  });

  it("returns warning for closing", () => {
    expect(posSessionStateTone("closing")).toBe("warning");
  });

  it("returns neutral for closed", () => {
    expect(posSessionStateTone("closed")).toBe("neutral");
  });
});

describe("posOrderStateTone", () => {
  it("returns success for done", () => {
    expect(posOrderStateTone("done")).toBe("success");
  });

  it("returns danger for refunded", () => {
    expect(posOrderStateTone("refunded")).toBe("danger");
  });
});

describe("session state guards", () => {
  it("canCloseSession only when opened", () => {
    expect(canCloseSession("opened")).toBe(true);
    expect(canCloseSession("closing")).toBe(false);
    expect(canCloseSession("closed")).toBe(false);
  });

  it("canStartClosing only when opened", () => {
    expect(canStartClosing("opened")).toBe(true);
    expect(canStartClosing("closing")).toBe(false);
    expect(canStartClosing("closed")).toBe(false);
  });

  it("canPlaceOrder only when opened", () => {
    expect(canPlaceOrder("opened")).toBe(true);
    expect(canPlaceOrder("closing")).toBe(false);
    expect(canPlaceOrder("closed")).toBe(false);
  });
});

describe("order state guards", () => {
  it("canInvoiceOrder when done and no invoice", () => {
    expect(canInvoiceOrder({ state: "done", invoiceId: null } as never)).toBe(true);
    expect(canInvoiceOrder({ state: "done", invoiceId: 1 } as never)).toBe(false);
    expect(canInvoiceOrder({ state: "refunded", invoiceId: null } as never)).toBe(false);
  });

  it("canRefundOrder when done", () => {
    expect(canRefundOrder({ state: "done" } as never)).toBe(true);
    expect(canRefundOrder({ state: "refunded" } as never)).toBe(false);
  });
});

describe("orderLineSubtotal", () => {
  it("computes qty * unitPrice without discount", () => {
    expect(orderLineSubtotal({ qty: 3, unitPrice: 100, discountPct: 0 } as never)).toBe(300);
  });

  it("applies percentage discount", () => {
    expect(orderLineSubtotal({ qty: 2, unitPrice: 100, discountPct: 10 } as never)).toBe(180);
  });

  it("handles zero qty", () => {
    expect(orderLineSubtotal({ qty: 0, unitPrice: 50, discountPct: 0 } as never)).toBe(0);
  });
});

describe("orderTotal", () => {
  it("sums line subtotals", () => {
    const lines = [
      { qty: 2, unitPrice: 100, discountPct: 0 },
      { qty: 1, unitPrice: 50, discountPct: 10 },
    ] as never[];
    expect(orderTotal(lines)).toBe(245);
  });

  it("returns 0 for empty lines", () => {
    expect(orderTotal([])).toBe(0);
  });
});

describe("orderTaxTotal", () => {
  it("sums tax amounts", () => {
    const lines = [{ priceTax: 20 }, { priceTax: 10 }] as never[];
    expect(orderTaxTotal(lines)).toBe(30);
  });
});

describe("sessionOrderTotal", () => {
  it("sums done orders only", () => {
    const orders = [
      { state: "done", amountTotal: 100 },
      { state: "refunded", amountTotal: 50 },
      { state: "done", amountTotal: 75 },
    ] as never[];
    expect(sessionOrderTotal(orders)).toBe(175);
  });

  it("returns 0 when no done orders", () => {
    expect(sessionOrderTotal([{ state: "refunded", amountTotal: 50 }] as never[])).toBe(0);
  });
});

describe("sessionPaymentSummary", () => {
  it("aggregates by method", () => {
    const payments = [
      { method: "cash", amount: 100 },
      { method: "card", amount: 50 },
      { method: "cash", amount: 25 },
    ];
    const result = sessionPaymentSummary(payments);
    expect(result).toEqual([
      { method: "cash", total: 125 },
      { method: "card", total: 50 },
    ]);
  });

  it("returns empty for no payments", () => {
    expect(sessionPaymentSummary([])).toEqual([]);
  });
});

describe("configDisplayName", () => {
  it("returns name when present", () => {
    expect(configDisplayName({ id: 1, name: "Main Store" } as never)).toBe("Main Store");
  });

  it("falls back to POS-{id}", () => {
    expect(configDisplayName({ id: 42, name: "" } as never)).toBe("POS-42");
  });
});
