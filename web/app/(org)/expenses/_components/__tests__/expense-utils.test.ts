import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canApprove,
  canPost,
  canRefuse,
  canReimburse,
  canSubmit,
  expenseStateTone,
} from "../expense-utils";

describe("expense-utils", () => {
  describe("canSubmit", () => {
    it("should allow submitting draft expenses", () => {
      expect(canSubmit("draft")).toBe(true);
    });

    it("should not allow submitting non-draft expenses", () => {
      expect(canSubmit("submitted")).toBe(false);
      expect(canSubmit("approved")).toBe(false);
      expect(canSubmit("refused")).toBe(false);
      expect(canSubmit("posted")).toBe(false);
      expect(canSubmit("reimbursed")).toBe(false);
    });
  });

  describe("canApprove", () => {
    it("should allow approving submitted expenses", () => {
      expect(canApprove("submitted")).toBe(true);
    });

    it("should not allow approving non-submitted expenses", () => {
      expect(canApprove("draft")).toBe(false);
      expect(canApprove("approved")).toBe(false);
      expect(canApprove("refused")).toBe(false);
      expect(canApprove("posted")).toBe(false);
      expect(canApprove("reimbursed")).toBe(false);
    });
  });

  describe("canRefuse", () => {
    it("should allow refusing submitted expenses", () => {
      expect(canRefuse("submitted")).toBe(true);
    });

    it("should not allow refusing non-submitted expenses", () => {
      expect(canRefuse("draft")).toBe(false);
      expect(canRefuse("approved")).toBe(false);
      expect(canRefuse("refused")).toBe(false);
      expect(canRefuse("posted")).toBe(false);
      expect(canRefuse("reimbursed")).toBe(false);
    });
  });

  describe("canPost", () => {
    it("should allow posting approved expenses", () => {
      expect(canPost("approved")).toBe(true);
    });

    it("should not allow posting non-approved expenses", () => {
      expect(canPost("draft")).toBe(false);
      expect(canPost("submitted")).toBe(false);
      expect(canPost("refused")).toBe(false);
      expect(canPost("posted")).toBe(false);
      expect(canPost("reimbursed")).toBe(false);
    });
  });

  describe("canReimburse", () => {
    it("should allow reimbursing posted expenses", () => {
      expect(canReimburse("posted")).toBe(true);
    });

    it("should not allow reimbursing non-posted expenses", () => {
      expect(canReimburse("draft")).toBe(false);
      expect(canReimburse("submitted")).toBe(false);
      expect(canReimburse("approved")).toBe(false);
      expect(canReimburse("refused")).toBe(false);
      expect(canReimburse("reimbursed")).toBe(false);
    });
  });

  describe("expenseStateTone", () => {
    it("should return neutral for draft", () => {
      expect(expenseStateTone("draft")).toBe("neutral");
    });

    it("should return info for submitted", () => {
      expect(expenseStateTone("submitted")).toBe("info");
    });

    it("should return success for approved", () => {
      expect(expenseStateTone("approved")).toBe("success");
    });

    it("should return danger for refused", () => {
      expect(expenseStateTone("refused")).toBe("danger");
    });

    it("should return warning for posted", () => {
      expect(expenseStateTone("posted")).toBe("warning");
    });

    it("should return success for reimbursed", () => {
      expect(expenseStateTone("reimbursed")).toBe("success");
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

  describe("formatCurrency", () => {
    it("should format as currency", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });
  });
});
