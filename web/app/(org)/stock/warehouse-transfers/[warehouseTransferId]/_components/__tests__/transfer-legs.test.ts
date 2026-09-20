import { describe, expect, it } from "vitest";

type TransferState = "draft" | "sent" | "in_transit" | "received" | "cancelled";

const transferSteps = ["draft", "sent", "in_transit", "received"] as const;

function transferStateIndex(state: TransferState): number {
  const idx = transferSteps.indexOf(state as (typeof transferSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function canSend(state: TransferState): boolean {
  return state === "draft";
}

function canReceive(state: TransferState): boolean {
  return state === "in_transit";
}

describe("transfer state-gated actions", () => {
  it("allows send in draft state", () => {
    expect(canSend("draft")).toBe(true);
  });

  it("disallows send in sent state", () => {
    expect(canSend("sent")).toBe(false);
  });

  it("disallows send in in_transit state", () => {
    expect(canSend("in_transit")).toBe(false);
  });

  it("disallows send in received state", () => {
    expect(canSend("received")).toBe(false);
  });

  it("allows receive only in in_transit state", () => {
    expect(canReceive("in_transit")).toBe(true);
  });

  it("disallows receive in draft state", () => {
    expect(canReceive("draft")).toBe(false);
  });

  it("disallows receive in sent state", () => {
    expect(canReceive("sent")).toBe(false);
  });

  it("disallows receive in received state", () => {
    expect(canReceive("received")).toBe(false);
  });
});

describe("transfer state index", () => {
  it("returns 0 for draft", () => {
    expect(transferStateIndex("draft")).toBe(0);
  });

  it("returns 1 for sent", () => {
    expect(transferStateIndex("sent")).toBe(1);
  });

  it("returns 2 for in_transit", () => {
    expect(transferStateIndex("in_transit")).toBe(2);
  });

  it("returns 3 for received", () => {
    expect(transferStateIndex("received")).toBe(3);
  });

  it("returns 0 for cancelled (unknown step)", () => {
    expect(transferStateIndex("cancelled")).toBe(0);
  });
});

describe("transfer leg highlights", () => {
  function activeLeg(state: TransferState): string {
    if (state === "draft") return "out";
    if (state === "sent" || state === "in_transit") return "transit";
    if (state === "received") return "in";
    return "none";
  }

  it("highlights out leg in draft", () => {
    expect(activeLeg("draft")).toBe("out");
  });

  it("highlights transit leg in sent", () => {
    expect(activeLeg("sent")).toBe("transit");
  });

  it("highlights transit leg in in_transit", () => {
    expect(activeLeg("in_transit")).toBe("transit");
  });

  it("highlights in leg in received", () => {
    expect(activeLeg("received")).toBe("in");
  });

  it("highlights nothing for cancelled", () => {
    expect(activeLeg("cancelled")).toBe("none");
  });
});
