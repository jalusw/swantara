import { describe, expect, it } from "vitest";
import {
  canApprove,
  canCancel,
  canConfirm,
  canCreateQuoteRequest,
  canEdit,
  requisitionStateTone,
  requisitionTotalQty,
} from "../purchase-request-utils";

describe("purchase-request-utils", () => {
  describe("canConfirm", () => {
    it("returns true for draft", () => {
      expect(canConfirm("draft")).toBe(true);
    });

    it("returns false for confirmed", () => {
      expect(canConfirm("confirmed")).toBe(false);
    });

    it("returns false for approved", () => {
      expect(canConfirm("approved")).toBe(false);
    });
  });

  describe("canApprove", () => {
    it("returns true for confirmed", () => {
      expect(canApprove("confirmed")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canApprove("draft")).toBe(false);
    });

    it("returns false for approved", () => {
      expect(canApprove("approved")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("returns true for draft", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("returns true for confirmed", () => {
      expect(canCancel("confirmed")).toBe(true);
    });

    it("returns false for approved", () => {
      expect(canCancel("approved")).toBe(false);
    });

    it("returns false for done", () => {
      expect(canCancel("done")).toBe(false);
    });
  });

  describe("canCreateQuoteRequest", () => {
    it("returns true for approved", () => {
      expect(canCreateQuoteRequest("approved")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canCreateQuoteRequest("draft")).toBe(false);
    });

    it("returns false for confirmed", () => {
      expect(canCreateQuoteRequest("confirmed")).toBe(false);
    });
  });

  describe("canEdit", () => {
    it("returns true for draft", () => {
      expect(canEdit("draft")).toBe(true);
    });

    it("returns false for confirmed", () => {
      expect(canEdit("confirmed")).toBe(false);
    });
  });

  describe("requisitionStateTone", () => {
    it("returns neutral for draft", () => {
      expect(requisitionStateTone("draft")).toBe("neutral");
    });

    it("returns info for confirmed", () => {
      expect(requisitionStateTone("confirmed")).toBe("info");
    });

    it("returns success for approved", () => {
      expect(requisitionStateTone("approved")).toBe("success");
    });

    it("returns danger for cancelled", () => {
      expect(requisitionStateTone("cancelled")).toBe("danger");
    });
  });

  describe("requisitionTotalQty", () => {
    it("sums line quantities", () => {
      expect(requisitionTotalQty([{ qty: 5 }, { qty: 10 }, { qty: 3 }])).toBe(18);
    });

    it("returns 0 for empty lines", () => {
      expect(requisitionTotalQty([])).toBe(0);
    });
  });
});
