import { describe, expect, it } from "vitest";
import type { Payslip } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canClose,
  canConfirm,
  canEdit,
  canPay,
  payrollRunStateTone,
  totalGross,
  totalNet,
} from "../payroll-utils";

describe("payroll-utils", () => {
  describe("canConfirm", () => {
    it("should allow confirming draft runs", () => {
      expect(canConfirm("draft")).toBe(true);
    });

    it("should not allow confirming non-draft runs", () => {
      expect(canConfirm("confirmed")).toBe(false);
      expect(canConfirm("paid")).toBe(false);
      expect(canConfirm("closed")).toBe(false);
    });
  });

  describe("canPay", () => {
    it("should allow paying confirmed runs", () => {
      expect(canPay("confirmed")).toBe(true);
    });

    it("should not allow paying non-confirmed runs", () => {
      expect(canPay("draft")).toBe(false);
      expect(canPay("paid")).toBe(false);
      expect(canPay("closed")).toBe(false);
    });
  });

  describe("canClose", () => {
    it("should allow closing paid runs", () => {
      expect(canClose("paid")).toBe(true);
    });

    it("should not allow closing non-paid runs", () => {
      expect(canClose("draft")).toBe(false);
      expect(canClose("confirmed")).toBe(false);
      expect(canClose("closed")).toBe(false);
    });
  });

  describe("canEdit", () => {
    it("should allow editing draft runs", () => {
      expect(canEdit("draft")).toBe(true);
    });

    it("should not allow editing non-draft runs", () => {
      expect(canEdit("confirmed")).toBe(false);
      expect(canEdit("paid")).toBe(false);
      expect(canEdit("closed")).toBe(false);
    });
  });

  describe("payrollRunStateTone", () => {
    it("should return neutral for draft", () => {
      expect(payrollRunStateTone("draft")).toBe("neutral");
    });

    it("should return info for confirmed", () => {
      expect(payrollRunStateTone("confirmed")).toBe("info");
    });

    it("should return warning for paid", () => {
      expect(payrollRunStateTone("paid")).toBe("warning");
    });

    it("should return success for closed", () => {
      expect(payrollRunStateTone("closed")).toBe("success");
    });
  });

  describe("totalGross", () => {
    it("should sum gross amounts", () => {
      const payslips = [
        { gross: 5000, net: 4500 },
        { gross: 3000, net: 2700 },
      ] as Payslip[];
      expect(totalGross(payslips)).toBe(8000);
    });

    it("should return 0 for empty list", () => {
      expect(totalGross([])).toBe(0);
    });
  });

  describe("totalNet", () => {
    it("should sum net amounts", () => {
      const payslips = [
        { gross: 5000, net: 4500 },
        { gross: 3000, net: 2700 },
      ] as Payslip[];
      expect(totalNet(payslips)).toBe(7200);
    });

    it("should return 0 for empty list", () => {
      expect(totalNet([])).toBe(0);
    });
  });

  describe("formatDate", () => {
    it("should format a valid date", () => {
      const result = formatDate("2024-03-15");
      expect(result).toContain("Mar");
      expect(result).toContain("15");
      expect(result).toContain("2024");
    });
  });

  describe("formatMoney", () => {
    it("should format as currency", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });
  });
});
