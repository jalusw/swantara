import { describe, expect, it } from "vitest";
import {
  canCancel,
  canConfirm,
  canDone,
  canReceive,
  canRefund,
  type Disposition,
  dispositionTone,
  rmaStateTone,
  rmaSteps,
  totalQty,
} from "../rma-utils";

describe("rma-utils", () => {
  describe("canConfirm", () => {
    it("returns true for draft state", () => {
      expect(canConfirm("draft")).toBe(true);
    });

    it("returns false for non-draft states", () => {
      expect(canConfirm("confirmed")).toBe(false);
      expect(canConfirm("received")).toBe(false);
      expect(canConfirm("refunded")).toBe(false);
      expect(canConfirm("done")).toBe(false);
      expect(canConfirm("cancelled")).toBe(false);
    });
  });

  describe("canReceive", () => {
    it("returns true for confirmed state", () => {
      expect(canReceive("confirmed")).toBe(true);
    });

    it("returns false for non-confirmed states", () => {
      expect(canReceive("draft")).toBe(false);
      expect(canReceive("received")).toBe(false);
      expect(canReceive("refunded")).toBe(false);
      expect(canReceive("done")).toBe(false);
      expect(canReceive("cancelled")).toBe(false);
    });
  });

  describe("canRefund", () => {
    it("returns true for received state", () => {
      expect(canRefund("received")).toBe(true);
    });

    it("returns false for non-received states", () => {
      expect(canRefund("draft")).toBe(false);
      expect(canRefund("confirmed")).toBe(false);
      expect(canRefund("refunded")).toBe(false);
      expect(canRefund("done")).toBe(false);
      expect(canRefund("cancelled")).toBe(false);
    });
  });

  describe("canDone", () => {
    it("returns true for refunded state", () => {
      expect(canDone("refunded")).toBe(true);
    });

    it("returns false for non-refunded states", () => {
      expect(canDone("draft")).toBe(false);
      expect(canDone("confirmed")).toBe(false);
      expect(canDone("received")).toBe(false);
      expect(canDone("done")).toBe(false);
      expect(canDone("cancelled")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("returns true for draft state", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("returns true for confirmed state", () => {
      expect(canCancel("confirmed")).toBe(true);
    });

    it("returns false for other states", () => {
      expect(canCancel("received")).toBe(false);
      expect(canCancel("refunded")).toBe(false);
      expect(canCancel("done")).toBe(false);
      expect(canCancel("cancelled")).toBe(false);
    });
  });

  describe("rmaStateTone", () => {
    it("returns neutral for draft", () => {
      expect(rmaStateTone("draft")).toBe("neutral");
    });

    it("returns info for confirmed", () => {
      expect(rmaStateTone("confirmed")).toBe("info");
    });

    it("returns info for received", () => {
      expect(rmaStateTone("received")).toBe("info");
    });

    it("returns success for refunded", () => {
      expect(rmaStateTone("refunded")).toBe("success");
    });

    it("returns success for done", () => {
      expect(rmaStateTone("done")).toBe("success");
    });

    it("returns danger for cancelled", () => {
      expect(rmaStateTone("cancelled")).toBe("danger");
    });
  });

  describe("dispositionTone", () => {
    it("returns success for restock", () => {
      expect(dispositionTone("restock")).toBe("success");
    });

    it("returns danger for scrap", () => {
      expect(dispositionTone("scrap")).toBe("danger");
    });

    it("returns warning for repair", () => {
      expect(dispositionTone("repair")).toBe("warning");
    });

    it("returns info for replace", () => {
      expect(dispositionTone("replace")).toBe("info");
    });
  });

  describe("rmaSteps", () => {
    it("returns all steps up to draft", () => {
      expect(rmaSteps("draft")).toEqual(["draft"]);
    });

    it("returns all steps up to confirmed", () => {
      expect(rmaSteps("confirmed")).toEqual(["draft", "confirmed"]);
    });

    it("returns all steps up to received", () => {
      expect(rmaSteps("received")).toEqual(["draft", "confirmed", "received"]);
    });

    it("returns all steps up to refunded", () => {
      expect(rmaSteps("refunded")).toEqual(["draft", "confirmed", "received", "refunded"]);
    });

    it("returns all steps up to done", () => {
      expect(rmaSteps("done")).toEqual(["draft", "confirmed", "received", "refunded", "done"]);
    });

    it("returns cancelled steps for cancelled state", () => {
      expect(rmaSteps("cancelled")).toEqual(["draft", "cancelled"]);
    });
  });

  describe("totalQty", () => {
    it("returns 0 for empty lines", () => {
      expect(totalQty([])).toBe(0);
    });

    it("sums quantities from lines", () => {
      const lines = [
        {
          id: 1,
          rmaId: 1,
          itemId: 1,
          qty: 5,
          batchId: null,
          disposition: "restock" as Disposition,
          stockMovementId: null,
          creditNoteId: null,
        },
        {
          id: 2,
          rmaId: 1,
          itemId: 2,
          qty: 3,
          batchId: null,
          disposition: "scrap" as Disposition,
          stockMovementId: null,
          creditNoteId: null,
        },
        {
          id: 3,
          rmaId: 1,
          itemId: 3,
          qty: 2,
          batchId: null,
          disposition: "repair" as Disposition,
          stockMovementId: null,
          creditNoteId: null,
        },
      ];
      expect(totalQty(lines)).toBe(10);
    });
  });
});
