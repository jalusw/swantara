import { describe, expect, it } from "vitest";
import {
  canCancelMo,
  canConfirmMo,
  canPlanMo,
  canProduceMo,
  canSettleMo,
  canStartMo,
  componentConsumptionProgress,
  moProgress,
  moStateLabel,
  moStateTone,
} from "../production-order-utils";

describe("production-order-utils", () => {
  describe("moStateLabel", () => {
    it("maps all states to translation keys", () => {
      expect(moStateLabel("draft")).toBe("Draft");
      expect(moStateLabel("confirmed")).toBe("Confirmed");
      expect(moStateLabel("planned")).toBe("Planned");
      expect(moStateLabel("in_progress")).toBe("In Progress");
      expect(moStateLabel("done")).toBe("Done");
      expect(moStateLabel("cancelled")).toBe("Cancelled");
    });
  });

  describe("canConfirmMo", () => {
    it("allows confirm only in draft", () => {
      expect(canConfirmMo("draft")).toBe(true);
      expect(canConfirmMo("confirmed")).toBe(false);
    });
  });

  describe("canPlanMo", () => {
    it("allows plan only in confirmed", () => {
      expect(canPlanMo("confirmed")).toBe(true);
      expect(canPlanMo("planned")).toBe(false);
    });
  });

  describe("canStartMo", () => {
    it("allows start only in planned", () => {
      expect(canStartMo("planned")).toBe(true);
      expect(canStartMo("in_progress")).toBe(false);
    });
  });

  describe("canProduceMo", () => {
    it("allows produce only in in_progress", () => {
      expect(canProduceMo("in_progress")).toBe(true);
      expect(canProduceMo("done")).toBe(false);
    });
  });

  describe("canSettleMo", () => {
    it("allows settle only in in_progress", () => {
      expect(canSettleMo("in_progress")).toBe(true);
      expect(canSettleMo("done")).toBe(false);
    });
  });

  describe("canCancelMo", () => {
    it("allows cancel in all states except done and cancelled", () => {
      expect(canCancelMo("draft")).toBe(true);
      expect(canCancelMo("confirmed")).toBe(true);
      expect(canCancelMo("planned")).toBe(true);
      expect(canCancelMo("in_progress")).toBe(true);
      expect(canCancelMo("done")).toBe(false);
      expect(canCancelMo("cancelled")).toBe(false);
    });
  });

  describe("moStateTone", () => {
    it("returns correct tone for each state", () => {
      expect(moStateTone("draft")).toBe("neutral");
      expect(moStateTone("confirmed")).toBe("info");
      expect(moStateTone("planned")).toBe("info");
      expect(moStateTone("in_progress")).toBe("warning");
      expect(moStateTone("done")).toBe("success");
      expect(moStateTone("cancelled")).toBe("danger");
    });
  });

  describe("moProgress", () => {
    it("calculates progress percentage", () => {
      expect(moProgress({ qtyToProduce: 100, qtyProduced: 50 })).toBe(50);
      expect(moProgress({ qtyToProduce: 0, qtyProduced: 0 })).toBe(0);
      expect(moProgress({ qtyToProduce: 100, qtyProduced: 150 })).toBe(100);
    });
  });

  describe("componentConsumptionProgress", () => {
    it("calculates consumption progress", () => {
      expect(
        componentConsumptionProgress({
          qtyPlanned: 10,
          qtyConsumed: 5,
        }),
      ).toBe(50);
      expect(
        componentConsumptionProgress({
          qtyPlanned: 0,
          qtyConsumed: 0,
        }),
      ).toBe(0);
    });
  });
});
