import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canActivate,
  canChurn,
  canClose,
  canPause,
  canResume,
  formatInterval,
  subscriptionStateTone,
} from "../subscription-utils";

describe("subscription-utils", () => {
  describe("canActivate", () => {
    it("should allow activating draft subscriptions", () => {
      expect(canActivate("draft")).toBe(true);
    });

    it("should not allow activating non-draft subscriptions", () => {
      expect(canActivate("active")).toBe(false);
      expect(canActivate("paused")).toBe(false);
      expect(canActivate("churned")).toBe(false);
      expect(canActivate("closed")).toBe(false);
    });
  });

  describe("canPause", () => {
    it("should allow pausing active subscriptions", () => {
      expect(canPause("active")).toBe(true);
    });

    it("should not allow pausing non-active subscriptions", () => {
      expect(canPause("draft")).toBe(false);
      expect(canPause("paused")).toBe(false);
      expect(canPause("churned")).toBe(false);
      expect(canPause("closed")).toBe(false);
    });
  });

  describe("canResume", () => {
    it("should allow resuming paused subscriptions", () => {
      expect(canResume("paused")).toBe(true);
    });

    it("should not allow resuming non-paused subscriptions", () => {
      expect(canResume("draft")).toBe(false);
      expect(canResume("active")).toBe(false);
      expect(canResume("churned")).toBe(false);
      expect(canResume("closed")).toBe(false);
    });
  });

  describe("canChurn", () => {
    it("should allow churning active or paused subscriptions", () => {
      expect(canChurn("active")).toBe(true);
      expect(canChurn("paused")).toBe(true);
    });

    it("should not allow churning other states", () => {
      expect(canChurn("draft")).toBe(false);
      expect(canChurn("churned")).toBe(false);
      expect(canChurn("closed")).toBe(false);
    });
  });

  describe("canClose", () => {
    it("should allow closing active or paused subscriptions", () => {
      expect(canClose("active")).toBe(true);
      expect(canClose("paused")).toBe(true);
    });

    it("should not allow closing other states", () => {
      expect(canClose("draft")).toBe(false);
      expect(canClose("churned")).toBe(false);
      expect(canClose("closed")).toBe(false);
    });
  });

  describe("subscriptionStateTone", () => {
    it("should return neutral for draft", () => {
      expect(subscriptionStateTone("draft")).toBe("neutral");
    });

    it("should return success for active", () => {
      expect(subscriptionStateTone("active")).toBe("success");
    });

    it("should return warning for paused", () => {
      expect(subscriptionStateTone("paused")).toBe("warning");
    });

    it("should return danger for churned", () => {
      expect(subscriptionStateTone("churned")).toBe("danger");
    });

    it("should return info for closed", () => {
      expect(subscriptionStateTone("closed")).toBe("info");
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

  describe("formatInterval", () => {
    it("should format single count", () => {
      expect(formatInterval("monthly", 1)).toBe("monthly");
    });

    it("should format multiple count", () => {
      expect(formatInterval("monthly", 3)).toBe("monthly (x3)");
    });
  });
});
