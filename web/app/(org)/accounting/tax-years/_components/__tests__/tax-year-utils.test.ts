import { describe, expect, it } from "vitest";
import type { TaxYear } from "@/lib/services/swantara";
import { isPeriodLocked, isYearLocked, yearDurationMonths } from "../tax-year-utils";

function taxYear(overrides: Partial<TaxYear>): TaxYear {
  return {
    id: overrides.id ?? 1,
    createdAt: new Date(),
    updatedAt: new Date(),
    organizationId: 1,
    name: overrides.name ?? "FY 2026",
    dateStart: overrides.dateStart ?? null,
    dateEnd: overrides.dateEnd ?? null,
    state: overrides.state ?? "open",
    ...overrides,
  } as TaxYear;
}

describe("isYearLocked", () => {
  it("returns true when state is done", () => {
    expect(isYearLocked(taxYear({ state: "done" }))).toBe(true);
  });

  it("returns true when state is locked", () => {
    expect(isYearLocked(taxYear({ state: "locked" }))).toBe(true);
  });

  it("returns false when state is open", () => {
    expect(isYearLocked(taxYear({ state: "open" }))).toBe(false);
  });
});

describe("isPeriodLocked", () => {
  it("returns true when year state is done", () => {
    const year = taxYear({
      state: "done",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2026-06-15")).toBe(true);
  });

  it("returns true when year state is locked", () => {
    const year = taxYear({
      state: "locked",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2026-06-15")).toBe(true);
  });

  it("returns true when date is before year start", () => {
    const year = taxYear({
      state: "open",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2025-12-31")).toBe(true);
  });

  it("returns true when date is after year end", () => {
    const year = taxYear({
      state: "open",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2027-01-01")).toBe(true);
  });

  it("returns false when date is within open year", () => {
    const year = taxYear({
      state: "open",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2026-06-15")).toBe(false);
  });

  it("returns false when date is on year boundaries", () => {
    const year = taxYear({
      state: "open",
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(isPeriodLocked(year, "2026-01-01")).toBe(false);
    expect(isPeriodLocked(year, "2026-12-31")).toBe(false);
  });
});

describe("yearDurationMonths", () => {
  it("returns 12 for a full calendar year", () => {
    const year = taxYear({
      dateStart: new Date("2026-01-01"),
      dateEnd: new Date("2026-12-31"),
    });
    expect(yearDurationMonths(year)).toBe(12);
  });

  it("returns correct months for partial year", () => {
    const year = taxYear({
      dateStart: new Date("2026-04-01"),
      dateEnd: new Date("2026-09-30"),
    });
    expect(yearDurationMonths(year)).toBe(6);
  });

  it("returns 0 when start date is null", () => {
    const year = taxYear({ dateStart: null, dateEnd: new Date("2026-12-31") });
    expect(yearDurationMonths(year)).toBe(0);
  });

  it("returns 0 when end date is null", () => {
    const year = taxYear({ dateStart: new Date("2026-01-01"), dateEnd: null });
    expect(yearDurationMonths(year)).toBe(0);
  });

  it("returns 0 when both dates are null", () => {
    const year = taxYear({ dateStart: null, dateEnd: null });
    expect(yearDurationMonths(year)).toBe(0);
  });

  it("returns 1 for same month start and end", () => {
    const year = taxYear({
      dateStart: new Date("2026-06-01"),
      dateEnd: new Date("2026-06-30"),
    });
    expect(yearDurationMonths(year)).toBe(1);
  });
});
