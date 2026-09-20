import { describe, expect, it } from "vitest";
import type { AssetDepreciationLine } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canDispose,
  canGenerateSchedule,
  canPostDepreciation,
  fixedAssetStateTone,
  nbv,
  totalDepreciation,
} from "../fixed-asset-utils";

describe("fixed-asset-utils", () => {
  describe("canGenerateSchedule", () => {
    it("should allow generating schedule for running assets", () => {
      expect(canGenerateSchedule("running")).toBe(true);
    });

    it("should not allow generating schedule for other states", () => {
      expect(canGenerateSchedule("draft")).toBe(false);
      expect(canGenerateSchedule("sold")).toBe(false);
      expect(canGenerateSchedule("disposed")).toBe(false);
    });
  });

  describe("canPostDepreciation", () => {
    const lines = [
      { posted: true, accumulated: 1000 },
      { posted: false, accumulated: 500 },
    ] as AssetDepreciationLine[];

    it("should allow posting when running and has unposted lines", () => {
      expect(canPostDepreciation("running", lines)).toBe(true);
    });

    it("should not allow posting when all lines are posted", () => {
      const allPosted = lines.map((l) => ({ ...l, posted: true }));
      expect(canPostDepreciation("running", allPosted)).toBe(false);
    });

    it("should not allow posting when not running", () => {
      expect(canPostDepreciation("draft", lines)).toBe(false);
      expect(canPostDepreciation("sold", lines)).toBe(false);
      expect(canPostDepreciation("disposed", lines)).toBe(false);
    });
  });

  describe("canDispose", () => {
    it("should allow disposing running assets", () => {
      expect(canDispose("running")).toBe(true);
    });

    it("should not allow disposing other states", () => {
      expect(canDispose("draft")).toBe(false);
      expect(canDispose("sold")).toBe(false);
      expect(canDispose("disposed")).toBe(false);
    });
  });

  describe("fixedAssetStateTone", () => {
    it("should return neutral for draft", () => {
      expect(fixedAssetStateTone("draft")).toBe("neutral");
    });

    it("should return info for running", () => {
      expect(fixedAssetStateTone("running")).toBe("info");
    });

    it("should return success for sold", () => {
      expect(fixedAssetStateTone("sold")).toBe("success");
    });

    it("should return warning for disposed", () => {
      expect(fixedAssetStateTone("disposed")).toBe("warning");
    });
  });

  describe("nbv", () => {
    it("should calculate net book value", () => {
      const lines = [
        { posted: true, accumulated: 1000 },
        { posted: true, accumulated: 500 },
        { posted: false, accumulated: 300 },
      ] as AssetDepreciationLine[];
      expect(nbv(5000, lines)).toBe(3500);
    });

    it("should return full purchase value when no posted lines", () => {
      expect(nbv(5000, [])).toBe(5000);
    });

    it("should ignore unposted lines", () => {
      const lines = [{ posted: false, accumulated: 2000 }] as AssetDepreciationLine[];
      expect(nbv(5000, lines)).toBe(5000);
    });
  });

  describe("totalDepreciation", () => {
    it("should sum posted amounts", () => {
      const lines = [
        { posted: true, amount: 1000 },
        { posted: false, amount: 500 },
        { posted: true, amount: 300 },
      ] as AssetDepreciationLine[];
      expect(totalDepreciation(lines)).toBe(1300);
    });

    it("should return 0 for empty list", () => {
      expect(totalDepreciation([])).toBe(0);
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
