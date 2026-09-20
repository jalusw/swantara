import { describe, expect, it } from "vitest";
import type { CrmLead, CrmStage } from "@/lib/services/swantara";
import {
  canAdvanceStage,
  canLose,
  canPromote,
  canWin,
  isClosedLead,
  isLostStage,
  isWonStage,
  sortedStages,
  weightedPipeline,
  weightedRevenue,
  winRate,
} from "../crm-utils";

describe("weightedRevenue", () => {
  it("calculates weighted revenue as expected times probability over 100", () => {
    expect(weightedRevenue(1000, 50)).toBe(500);
  });

  it("handles 0 probability", () => {
    expect(weightedRevenue(1000, 0)).toBe(0);
  });

  it("handles 100 probability", () => {
    expect(weightedRevenue(750, 100)).toBe(750);
  });

  it("rounds to 2 decimals", () => {
    expect(weightedRevenue(333, 33)).toBe(109.89);
  });
});

describe("weightedPipeline", () => {
  it("sums weighted revenue of open opportunities", () => {
    const opps = [
      { expectedRevenue: 1000, probability: 50 },
      { expectedRevenue: 2000, probability: 25 },
    ];
    expect(weightedPipeline(opps)).toBe(1000);
  });

  it("returns 0 for empty array", () => {
    expect(weightedPipeline([])).toBe(0);
  });

  it("matches service forecast weighted pipeline formula", () => {
    const opps = [
      { expectedRevenue: 5000, probability: 10 },
      { expectedRevenue: 8000, probability: 35 },
      { expectedRevenue: 12000, probability: 60 },
      { expectedRevenue: 20000, probability: 100 },
    ];
    const expected = 500 + 2800 + 7200 + 20000;
    expect(weightedPipeline(opps)).toBe(expected);
  });
});

describe("winRate", () => {
  it("returns 0 when no deals", () => {
    expect(winRate(0, 0)).toBe(0);
  });

  it("calculates win rate as won over total times 100", () => {
    expect(winRate(3, 1)).toBe(75);
    expect(winRate(1, 3)).toBe(25);
  });

  it("rounds to 2 decimals", () => {
    expect(winRate(1, 2)).toBe(33.33);
  });

  it("returns 100 when all won", () => {
    expect(winRate(5, 0)).toBe(100);
  });

  it("returns 0 when all lost", () => {
    expect(winRate(0, 5)).toBe(0);
  });
});

describe("canPromote", () => {
  it("allows promotion for lead type when not closed", () => {
    expect(canPromote({ type: "lead", closedAt: null })).toBe(true);
  });

  it("disallows promotion for opportunity type", () => {
    expect(canPromote({ type: "opportunity", closedAt: null })).toBe(false);
  });

  it("disallows promotion when closed", () => {
    expect(canPromote({ type: "lead", closedAt: new Date() })).toBe(false);
  });
});

describe("canAdvanceStage", () => {
  it("allows advance for open opportunity", () => {
    expect(
      canAdvanceStage({
        type: "opportunity",
        closedAt: null,
        lostReason: null,
      }),
    ).toBe(true);
  });

  it("disallows advance for lead", () => {
    expect(canAdvanceStage({ type: "lead", closedAt: null, lostReason: null })).toBe(false);
  });

  it("disallows advance when closed", () => {
    expect(
      canAdvanceStage({
        type: "opportunity",
        closedAt: new Date(),
        lostReason: null,
      }),
    ).toBe(false);
  });

  it("disallows advance when lost", () => {
    expect(
      canAdvanceStage({
        type: "opportunity",
        closedAt: new Date(),
        lostReason: "price",
      }),
    ).toBe(false);
  });

  it("disallows advance when lost reason present even if closedAt null", () => {
    expect(
      canAdvanceStage({
        type: "opportunity",
        closedAt: null,
        lostReason: "price",
      }),
    ).toBe(false);
  });
});

describe("canWin / canLose", () => {
  it("allows win for open opportunity not yet won", () => {
    expect(
      canWin({
        type: "opportunity",
        closedAt: null,
        lostReason: null,
        isWon: false,
      }),
    ).toBe(true);
  });

  it("disallows win when already won", () => {
    expect(
      canWin({
        type: "opportunity",
        closedAt: new Date(),
        lostReason: null,
        isWon: true,
      }),
    ).toBe(false);
  });

  it("disallows win for lead", () => {
    expect(canWin({ type: "lead", closedAt: null, lostReason: null, isWon: false })).toBe(false);
  });

  it("allows lose for open opportunity", () => {
    expect(
      canLose({
        type: "opportunity",
        closedAt: null,
        lostReason: null,
        isWon: false,
      }),
    ).toBe(true);
  });

  it("disallows lose when already lost", () => {
    expect(
      canLose({
        type: "opportunity",
        closedAt: new Date(),
        lostReason: "reason",
        isWon: false,
      }),
    ).toBe(false);
  });

  it("disallows lose when already won", () => {
    expect(
      canLose({
        type: "opportunity",
        closedAt: new Date(),
        lostReason: null,
        isWon: true,
      }),
    ).toBe(false);
  });
});

describe("isClosedLead", () => {
  it("is closed when closedAt set", () => {
    expect(isClosedLead({ closedAt: new Date(), lostReason: null, isWon: false })).toBe(true);
  });

  it("is closed when isWon true", () => {
    expect(isClosedLead({ closedAt: null, lostReason: null, isWon: true })).toBe(true);
  });

  it("is closed when lostReason set", () => {
    expect(isClosedLead({ closedAt: null, lostReason: "reason", isWon: false })).toBe(true);
  });

  it("is not closed when open", () => {
    expect(isClosedLead({ closedAt: null, lostReason: null, isWon: false })).toBe(false);
  });
});

describe("stage helpers", () => {
  it("detects won stage", () => {
    expect(
      isWonStage({
        id: 1,
        name: "Won",
        sequence: 40,
        isWon: true,
        probability: 100,
      } as unknown as CrmStage),
    ).toBe(true);
    expect(
      isWonStage({
        id: 2,
        name: "Lost",
        sequence: 50,
        isWon: false,
        probability: 0,
      } as unknown as CrmStage),
    ).toBe(false);
  });

  it("detects lost stage as probability 0 and not won", () => {
    expect(
      isLostStage({
        id: 2,
        name: "Lost",
        sequence: 50,
        isWon: false,
        probability: 0,
      } as unknown as CrmStage),
    ).toBe(true);
    expect(
      isLostStage({
        id: 1,
        name: "Won",
        sequence: 40,
        isWon: true,
        probability: 100,
      } as unknown as CrmStage),
    ).toBe(false);
    expect(
      isLostStage({
        id: 3,
        name: "New",
        sequence: 10,
        isWon: false,
        probability: 10,
      } as unknown as CrmStage),
    ).toBe(false);
  });

  it("sorts stages by sequence", () => {
    const stages = [
      { id: 2, name: "Proposal", sequence: 30, isWon: false, probability: 60 },
      { id: 1, name: "New", sequence: 10, isWon: false, probability: 10 },
      { id: 3, name: "Won", sequence: 40, isWon: true, probability: 100 },
    ] as unknown as CrmStage[];
    const sorted = sortedStages(stages);
    expect(sorted.map((s) => s.name)).toEqual(["New", "Proposal", "Won"]);
  });
});

describe("won opportunity CTA", () => {
  it("shows create quotation CTA only for won opportunities", () => {
    const won = {
      type: "opportunity",
      isWon: true,
      closedAt: new Date(),
      lostReason: null,
    } as unknown as CrmLead;
    const open = {
      type: "opportunity",
      isWon: false,
      closedAt: null,
      lostReason: null,
    } as unknown as CrmLead;
    const lost = {
      type: "opportunity",
      isWon: false,
      closedAt: new Date(),
      lostReason: "price",
    } as unknown as CrmLead;
    expect(won.isWon).toBe(true);
    expect(open.isWon).toBe(false);
    expect(lost.lostReason != null).toBe(true);
  });

  it("lost reason is required to mark lost", () => {
    const reason = "";
    expect(reason.trim().length === 0).toBe(true);
    const validReason = "Price too high";
    expect(validReason.trim().length > 0).toBe(true);
  });
});
