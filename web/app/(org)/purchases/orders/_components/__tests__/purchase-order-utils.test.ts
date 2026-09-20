import { describe, expect, it } from "vitest";
import {
  canBill,
  canCancel,
  canConfirm,
  canEdit,
  canPay,
  canReceive,
  orderRemainingToBill,
  orderRemainingToReceive,
  remainingToBill,
  remainingToReceive,
  threeWayMatchStatus,
} from "../purchase-order-utils";

describe("purchase-order-utils", () => {
  describe("canConfirm", () => {
    it("returns true for draft", () => {
      expect(canConfirm("draft")).toBe(true);
    });

    it("returns false for sent", () => {
      expect(canConfirm("sent")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("returns true for draft", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("returns true for sent", () => {
      expect(canCancel("sent")).toBe(true);
    });

    it("returns false for confirmed", () => {
      expect(canCancel("confirmed")).toBe(false);
    });

    it("returns false for done", () => {
      expect(canCancel("done")).toBe(false);
    });
  });

  describe("canReceive", () => {
    it("returns true for confirmed", () => {
      expect(canReceive("confirmed")).toBe(true);
    });

    it("returns true for done", () => {
      expect(canReceive("done")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canReceive("draft")).toBe(false);
    });
  });

  describe("canBill", () => {
    it("returns true for confirmed", () => {
      expect(canBill("confirmed")).toBe(true);
    });

    it("returns true for done", () => {
      expect(canBill("done")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canBill("draft")).toBe(false);
    });
  });

  describe("canPay", () => {
    it("returns true for confirmed", () => {
      expect(canPay("confirmed")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canPay("draft")).toBe(false);
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

  describe("remainingToReceive", () => {
    it("computes remaining from ordered minus received", () => {
      expect(remainingToReceive({ qtyOrdered: 10, qtyReceived: 3, qtyReturns: 0 })).toBe(7);
    });

    it("returns 0 when fully received", () => {
      expect(remainingToReceive({ qtyOrdered: 10, qtyReceived: 10, qtyReturns: 0 })).toBe(0);
    });

    it("accounts for returns", () => {
      expect(remainingToReceive({ qtyOrdered: 10, qtyReceived: 8, qtyReturns: 2 })).toBe(4);
    });

    it("never goes negative", () => {
      expect(remainingToReceive({ qtyOrdered: 5, qtyReceived: 10, qtyReturns: 0 })).toBe(0);
    });
  });

  describe("remainingToBill", () => {
    it("computes remaining from ordered minus billed", () => {
      expect(remainingToBill({ qtyOrdered: 10, qtyBilled: 4, qtyReturns: 0 })).toBe(6);
    });

    it("returns 0 when fully billed", () => {
      expect(remainingToBill({ qtyOrdered: 10, qtyBilled: 10, qtyReturns: 0 })).toBe(0);
    });

    it("accounts for returns", () => {
      expect(remainingToBill({ qtyOrdered: 10, qtyBilled: 8, qtyReturns: 2 })).toBe(4);
    });
  });

  describe("orderRemainingToReceive", () => {
    it("sums remaining across lines", () => {
      expect(
        orderRemainingToReceive([
          {
            id: 1,
            orderId: 1,
            sequence: 10,
            itemId: 1,
            description: null,
            qtyOrdered: 10,
            qtyReceived: 3,
            qtyBilled: 0,
            qtyReturns: 0,
            unitId: null,
            unitPrice: 0,
            discountPct: 0,
            taxIds: [],
            dimensionId: null,
            priceSubtotal: 0,
          },
          {
            id: 2,
            orderId: 1,
            sequence: 20,
            itemId: 2,
            description: null,
            qtyOrdered: 5,
            qtyReceived: 5,
            qtyBilled: 0,
            qtyReturns: 0,
            unitId: null,
            unitPrice: 0,
            discountPct: 0,
            taxIds: [],
            dimensionId: null,
            priceSubtotal: 0,
          },
        ]),
      ).toBe(7);
    });
  });

  describe("orderRemainingToBill", () => {
    it("sums remaining across lines", () => {
      expect(
        orderRemainingToBill([
          {
            id: 1,
            orderId: 1,
            sequence: 10,
            itemId: 1,
            description: null,
            qtyOrdered: 10,
            qtyReceived: 0,
            qtyBilled: 2,
            qtyReturns: 0,
            unitId: null,
            unitPrice: 0,
            discountPct: 0,
            taxIds: [],
            dimensionId: null,
            priceSubtotal: 0,
          },
          {
            id: 2,
            orderId: 1,
            sequence: 20,
            itemId: 2,
            description: null,
            qtyOrdered: 5,
            qtyReceived: 0,
            qtyBilled: 5,
            qtyReturns: 0,
            unitId: null,
            unitPrice: 0,
            discountPct: 0,
            taxIds: [],
            dimensionId: null,
            priceSubtotal: 0,
          },
        ]),
      ).toBe(8);
    });
  });

  describe("threeWayMatchStatus", () => {
    it("returns matched when fully received and billed", () => {
      expect(
        threeWayMatchStatus({
          id: 1,
          orderId: 1,
          sequence: 10,
          itemId: 1,
          description: null,
          qtyOrdered: 10,
          qtyReceived: 10,
          qtyBilled: 10,
          qtyReturns: 0,
          unitId: null,
          unitPrice: 0,
          discountPct: 0,
          taxIds: [],
          dimensionId: null,
          priceSubtotal: 0,
        }),
      ).toBe("matched");
    });

    it("returns partial when partially received", () => {
      expect(
        threeWayMatchStatus({
          id: 1,
          orderId: 1,
          sequence: 10,
          itemId: 1,
          description: null,
          qtyOrdered: 10,
          qtyReceived: 5,
          qtyBilled: 0,
          qtyReturns: 0,
          unitId: null,
          unitPrice: 0,
          discountPct: 0,
          taxIds: [],
          dimensionId: null,
          priceSubtotal: 0,
        }),
      ).toBe("partial");
    });

    it("returns over_received when received exceeds ordered", () => {
      expect(
        threeWayMatchStatus({
          id: 1,
          orderId: 1,
          sequence: 10,
          itemId: 1,
          description: null,
          qtyOrdered: 10,
          qtyReceived: 12,
          qtyBilled: 0,
          qtyReturns: 0,
          unitId: null,
          unitPrice: 0,
          discountPct: 0,
          taxIds: [],
          dimensionId: null,
          priceSubtotal: 0,
        }),
      ).toBe("over_received");
    });

    it("returns over_billed when billed exceeds ordered", () => {
      expect(
        threeWayMatchStatus({
          id: 1,
          orderId: 1,
          sequence: 10,
          itemId: 1,
          description: null,
          qtyOrdered: 10,
          qtyReceived: 5,
          qtyBilled: 12,
          qtyReturns: 0,
          unitId: null,
          unitPrice: 0,
          discountPct: 0,
          taxIds: [],
          dimensionId: null,
          priceSubtotal: 0,
        }),
      ).toBe("over_billed");
    });
  });
});
