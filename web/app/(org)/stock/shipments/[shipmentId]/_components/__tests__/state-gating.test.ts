import { describe, expect, it } from "vitest";

type ShipmentState = "draft" | "waiting" | "confirmed" | "assigned" | "done" | "cancelled";

function canValidate(state: ShipmentState): boolean {
  return state === "draft" || state === "confirmed";
}

function canDone(state: ShipmentState): boolean {
  return state === "assigned";
}

const shipmentSteps = ["draft", "waiting", "confirmed", "assigned", "done"] as const;

function shipmentStateIndex(state: ShipmentState): number {
  const idx = shipmentSteps.indexOf(state as (typeof shipmentSteps)[number]);
  return idx >= 0 ? idx : 0;
}

describe("shipment state-gated actions", () => {
  it("allows validate in draft state", () => {
    expect(canValidate("draft")).toBe(true);
  });

  it("allows validate in confirmed state", () => {
    expect(canValidate("confirmed")).toBe(true);
  });

  it("disallows validate in waiting state", () => {
    expect(canValidate("waiting")).toBe(false);
  });

  it("disallows validate in assigned state", () => {
    expect(canValidate("assigned")).toBe(false);
  });

  it("disallows validate in done state", () => {
    expect(canValidate("done")).toBe(false);
  });

  it("allows done only in assigned state", () => {
    expect(canDone("assigned")).toBe(true);
  });

  it("disallows done in draft state", () => {
    expect(canDone("draft")).toBe(false);
  });

  it("disallows done in confirmed state", () => {
    expect(canDone("confirmed")).toBe(false);
  });

  it("disallows done in done state", () => {
    expect(canDone("done")).toBe(false);
  });
});

describe("shipment state index", () => {
  it("returns 0 for draft", () => {
    expect(shipmentStateIndex("draft")).toBe(0);
  });

  it("returns 1 for waiting", () => {
    expect(shipmentStateIndex("waiting")).toBe(1);
  });

  it("returns 2 for confirmed", () => {
    expect(shipmentStateIndex("confirmed")).toBe(2);
  });

  it("returns 3 for assigned", () => {
    expect(shipmentStateIndex("assigned")).toBe(3);
  });

  it("returns 4 for done", () => {
    expect(shipmentStateIndex("done")).toBe(4);
  });

  it("returns 0 for cancelled (unknown step)", () => {
    expect(shipmentStateIndex("cancelled")).toBe(0);
  });
});
