import { describe, expect, it } from "vitest";
import { formatDate } from "@/lib/utils";
import { canActivate, canCancel, serviceContractStateTone } from "../service-contract-utils";

describe("service-contract-utils", () => {
  describe("canActivate", () => {
    it("should allow activating draft contracts", () => {
      expect(canActivate("draft")).toBe(true);
    });

    it("should not allow activating non-draft contracts", () => {
      expect(canActivate("active")).toBe(false);
      expect(canActivate("cancelled")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("should allow cancelling draft or active contracts", () => {
      expect(canCancel("draft")).toBe(true);
      expect(canCancel("active")).toBe(true);
    });

    it("should not allow cancelling cancelled contracts", () => {
      expect(canCancel("cancelled")).toBe(false);
    });
  });

  describe("serviceContractStateTone", () => {
    it("should return neutral for draft", () => {
      expect(serviceContractStateTone("draft")).toBe("neutral");
    });

    it("should return success for active", () => {
      expect(serviceContractStateTone("active")).toBe("success");
    });

    it("should return danger for cancelled", () => {
      expect(serviceContractStateTone("cancelled")).toBe("danger");
    });
  });

  describe("formatDate", () => {
    it("should format a valid date", () => {
      const result = formatDate(new Date("2024-03-15"));
      expect(result).toContain("Mar");
      expect(result).toContain("15");
      expect(result).toContain("2024");
    });

    it("should return dash for null", () => {
      expect(formatDate(null, { nullFallback: "—" })).toBe("—");
    });
  });
});
