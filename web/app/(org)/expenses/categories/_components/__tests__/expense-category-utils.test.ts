import { describe, expect, it } from "vitest";
import { formatAccountName } from "../expense-category-utils";

describe("expense-category-utils", () => {
  describe("formatAccountName", () => {
    it("should format code and name together", () => {
      expect(formatAccountName("6010", "Office Supplies")).toBe("6010 - Office Supplies");
    });

    it("should return only name when code is null", () => {
      expect(formatAccountName(null, "Office Supplies")).toBe("Office Supplies");
    });

    it("should return only code when name is null", () => {
      expect(formatAccountName("6010", null)).toBe("6010");
    });

    it("should return dash when both are null", () => {
      expect(formatAccountName(null, null)).toBe("—");
    });

    it("should return dash when both are empty strings", () => {
      expect(formatAccountName("", "")).toBe("—");
    });
  });
});
