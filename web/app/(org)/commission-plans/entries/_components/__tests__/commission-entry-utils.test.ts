import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canCancelEntry,
  canConfirm,
  canPay,
  commissionEntryStateTone,
} from "../commission-entry-utils";

describe("commission-entry-utils", () => {
  describe("canConfirm", () => {
    it("should allow confirming draft entries", () => {
      expect(canConfirm("draft")).toBe(true);
    });

    it("should not allow confirming non-draft entries", () => {
      expect(canConfirm("confirmed")).toBe(false);
      expect(canConfirm("paid")).toBe(false);
      expect(canConfirm("cancelled")).toBe(false);
    });
  });

  describe("canPay", () => {
    it("should allow paying confirmed entries", () => {
      expect(canPay("confirmed")).toBe(true);
    });

    it("should not allow paying non-confirmed entries", () => {
      expect(canPay("draft")).toBe(false);
      expect(canPay("paid")).toBe(false);
      expect(canPay("cancelled")).toBe(false);
    });
  });

  describe("canCancelEntry", () => {
    it("should allow cancelling draft or confirmed entries", () => {
      expect(canCancelEntry("draft")).toBe(true);
      expect(canCancelEntry("confirmed")).toBe(true);
    });

    it("should not allow cancelling other states", () => {
      expect(canCancelEntry("paid")).toBe(false);
      expect(canCancelEntry("cancelled")).toBe(false);
    });
  });

  describe("commissionEntryStateTone", () => {
    it("should return neutral for draft", () => {
      expect(commissionEntryStateTone("draft")).toBe("neutral");
    });

    it("should return info for confirmed", () => {
      expect(commissionEntryStateTone("confirmed")).toBe("info");
    });

    it("should return success for paid", () => {
      expect(commissionEntryStateTone("paid")).toBe("success");
    });

    it("should return danger for cancelled", () => {
      expect(commissionEntryStateTone("cancelled")).toBe("danger");
    });
  });

  describe("formatDate", () => {
    it("should format a valid date", () => {
      const result = formatDate(new Date("2024-03-15"), { nullFallback: "—" });
      expect(result).toContain("Mar");
      expect(result).toContain("15");
      expect(result).toContain("2024");
    });

    it("should return dash for null", () => {
      expect(formatDate(null, { nullFallback: "—" })).toBe("—");
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
