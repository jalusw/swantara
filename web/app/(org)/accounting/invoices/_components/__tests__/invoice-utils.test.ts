import { describe, expect, it } from "vitest";
import type { Invoice } from "@/lib/services/swantara";
import {
  canPay,
  canPostInvoice,
  invoiceRemaining,
  invoiceStateTone,
  paymentStateTone,
} from "../invoice-utils";

function invoice(overrides: Partial<Invoice>): Invoice {
  return {
    id: overrides.id ?? 1,
    createdAt: new Date(),
    updatedAt: new Date(),
    organizationId: 1,
    moveId: null,
    type: overrides.type ?? "customer_invoice",
    contactId: 1,
    name: null,
    reference: null,
    invoiceDate: null,
    dueDate: null,
    currencyCode: null,
    journalId: null,
    paymentTermId: null,
    state: overrides.state ?? "draft",
    paymentState: overrides.paymentState ?? "not_paid",
    amountUntaxed: 0,
    amountTax: 0,
    amountTotal: 0,
    amountResidual: overrides.amountResidual ?? 0,
    ...overrides,
  } as Invoice;
}

describe("invoiceRemaining", () => {
  it("returns amountResidual from invoice", () => {
    expect(invoiceRemaining(invoice({ amountResidual: 500 }))).toBe(500);
  });

  it("returns zero when fully paid", () => {
    expect(invoiceRemaining(invoice({ amountResidual: 0 }))).toBe(0);
  });
});

describe("canPostInvoice", () => {
  it("allows posting only in draft", () => {
    expect(canPostInvoice("draft")).toBe(true);
    expect(canPostInvoice("posted")).toBe(false);
    expect(canPostInvoice("cancelled")).toBe(false);
  });
});

describe("canPay", () => {
  it("allows pay when posted and not paid", () => {
    expect(canPay("posted", "not_paid")).toBe(true);
  });

  it("allows pay when posted and partially paid", () => {
    expect(canPay("posted", "partial")).toBe(true);
  });

  it("rejects pay when posted and fully paid", () => {
    expect(canPay("posted", "paid")).toBe(false);
  });

  it("rejects pay when draft", () => {
    expect(canPay("draft", "not_paid")).toBe(false);
  });

  it("rejects pay when cancelled", () => {
    expect(canPay("cancelled", "not_paid")).toBe(false);
  });

  it("rejects pay when in payment", () => {
    expect(canPay("posted", "in_payment")).toBe(false);
  });

  it("rejects pay when reversed", () => {
    expect(canPay("posted", "reversed")).toBe(false);
  });
});

describe("invoiceStateTone", () => {
  it("returns correct tone for each state", () => {
    expect(invoiceStateTone("draft")).toBe("neutral");
    expect(invoiceStateTone("posted")).toBe("info");
    expect(invoiceStateTone("cancelled")).toBe("danger");
  });
});

describe("paymentStateTone", () => {
  it("returns correct tone for each payment state", () => {
    expect(paymentStateTone("paid")).toBe("success");
    expect(paymentStateTone("partial")).toBe("warning");
    expect(paymentStateTone("not_paid")).toBe("danger");
    expect(paymentStateTone("in_payment")).toBe("info");
    expect(paymentStateTone("reversed")).toBe("danger");
  });
});
