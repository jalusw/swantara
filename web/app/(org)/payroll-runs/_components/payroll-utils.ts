import type { PayrollRun, Payslip } from "@/lib/services/swantara";

export type PayrollRunState = PayrollRun["state"];

export function canConfirm(state: PayrollRunState): boolean {
  return state === "draft";
}

export function canPay(state: PayrollRunState): boolean {
  return state === "confirmed";
}

export function canClose(state: PayrollRunState): boolean {
  return state === "paid";
}

export function canEdit(state: PayrollRunState): boolean {
  return state === "draft";
}

export function payrollRunStateTone(
  state: PayrollRunState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "confirmed":
      return "info";
    case "paid":
      return "warning";
    case "closed":
      return "success";
    default:
      return "neutral";
  }
}

export function totalGross(payslips: Payslip[]): number {
  return payslips.reduce((sum, p) => sum + p.gross, 0);
}

export function totalNet(payslips: Payslip[]): number {
  return payslips.reduce((sum, p) => sum + p.net, 0);
}
