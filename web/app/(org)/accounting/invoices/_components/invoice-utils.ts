import type { Invoice } from "@/lib/services/swantara";

export type InvoiceType = Invoice["type"];
export type InvoiceState = Invoice["state"];
export type PaymentState = Invoice["paymentState"];

export function invoiceTypeLabel(type: InvoiceType): string {
  switch (type) {
    case "customer_invoice":
      return "Customer Invoice";
    case "customer_credit_note":
      return "Customer Credit Note";
    case "vendor_bill":
      return "Supplier Bill";
    case "vendor_credit_note":
      return "Supplier Credit Note";
    default:
      return type;
  }
}

export function invoiceStateLabel(state: InvoiceState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "posted":
      return "Diposting";
    case "cancelled":
      return "Dibatalkan";
    default:
      return state;
  }
}

export function paymentStateLabel(state: PaymentState): string {
  switch (state) {
    case "not_paid":
      return "Not Paid";
    case "in_payment":
      return "In Payment";
    case "partial":
      return "Partially Paid";
    case "paid":
      return "Lunas";
    case "reversed":
      return "Reversed";
    default:
      return state;
  }
}

export function canPostInvoice(state: InvoiceState): boolean {
  return state === "draft";
}

export function canCancelInvoice(state: InvoiceState): boolean {
  return state === "draft";
}

export function canPay(state: InvoiceState, paymentState: PaymentState): boolean {
  return state === "posted" && (paymentState === "not_paid" || paymentState === "partial");
}

export function invoiceRemaining(invoice: Invoice): number {
  return invoice.amountResidual;
}

export function invoiceStateTone(
  state: InvoiceState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "posted":
      return "info";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function paymentStateTone(
  state: PaymentState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "paid":
      return "success";
    case "partial":
      return "warning";
    case "not_paid":
      return "danger";
    case "in_payment":
      return "info";
    case "reversed":
      return "danger";
    default:
      return "neutral";
  }
}
