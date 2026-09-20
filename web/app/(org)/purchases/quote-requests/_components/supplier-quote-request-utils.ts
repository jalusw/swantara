import type { SupplierQuoteRequest } from "@/lib/services/swantara";

export type QuoteRequestState = SupplierQuoteRequest["state"];

export function canSend(state: QuoteRequestState): boolean {
  return state === "draft";
}

export function canCancel(state: QuoteRequestState): boolean {
  return state === "draft" || state === "sent";
}

export function canCreatePo(state: QuoteRequestState): boolean {
  return state === "done";
}

export function canEdit(state: QuoteRequestState): boolean {
  return state === "draft";
}

export function canReceiveQuotes(state: QuoteRequestState): boolean {
  return state === "sent";
}

export function quoteRequestStateTone(
  state: QuoteRequestState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "sent":
      return "info";
    case "done":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}
