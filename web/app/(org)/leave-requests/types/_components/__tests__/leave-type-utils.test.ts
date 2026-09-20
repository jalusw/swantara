import { describe, expect, it } from "vitest";
import { formatDate } from "@/lib/utils";
import { canApprove, canRefuse, canSubmit, leaveRequestStateTone } from "../leave-type-utils";

describe("leave-type-utils", () => {
  describe("canSubmit", () => {
    it("should allow submitting draft requests", () => {
      expect(canSubmit("draft")).toBe(true);
    });

    it("should not allow submitting non-draft requests", () => {
      expect(canSubmit("submitted")).toBe(false);
      expect(canSubmit("approved")).toBe(false);
      expect(canSubmit("refused")).toBe(false);
    });
  });

  describe("canApprove", () => {
    it("should allow approving submitted requests", () => {
      expect(canApprove("submitted")).toBe(true);
    });

    it("should not allow approving non-submitted requests", () => {
      expect(canApprove("draft")).toBe(false);
      expect(canApprove("approved")).toBe(false);
      expect(canApprove("refused")).toBe(false);
    });
  });

  describe("canRefuse", () => {
    it("should allow refusing submitted requests", () => {
      expect(canRefuse("submitted")).toBe(true);
    });

    it("should not allow refusing non-submitted requests", () => {
      expect(canRefuse("draft")).toBe(false);
      expect(canRefuse("approved")).toBe(false);
      expect(canRefuse("refused")).toBe(false);
    });
  });

  describe("leaveRequestStateTone", () => {
    it("should return neutral for draft", () => {
      expect(leaveRequestStateTone("draft")).toBe("neutral");
    });

    it("should return info for submitted", () => {
      expect(leaveRequestStateTone("submitted")).toBe("info");
    });

    it("should return success for approved", () => {
      expect(leaveRequestStateTone("approved")).toBe("success");
    });

    it("should return danger for refused", () => {
      expect(leaveRequestStateTone("refused")).toBe("danger");
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
});
