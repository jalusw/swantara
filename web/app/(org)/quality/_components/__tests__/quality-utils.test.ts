import { describe, expect, it } from "vitest";
import type { QualityAlert, QualityCheck, QualityPoint } from "@/lib/services/swantara";
import {
  alertStateIndex,
  alertStateLabel,
  alertStateSteps,
  alertStateTone,
  alertTraceability,
  canRecordResult,
  canTransitionAlert,
  checkResultLabel,
  checkResultTone,
  checksPassRate,
  checksPendingCount,
  checksWithFailedCount,
  isMeasureType,
  isPassFailType,
  normRangeLabel,
  severityLabel,
  testTypeLabel,
} from "../quality-utils";

describe("quality-utils", () => {
  describe("checkResultLabel", () => {
    it("maps all results to translation keys", () => {
      expect(checkResultLabel("pending")).toBe("Pending");
      expect(checkResultLabel("pass")).toBe("Pass");
      expect(checkResultLabel("fail")).toBe("Fail");
    });
  });

  describe("alertStateLabel", () => {
    it("maps all states to translation keys", () => {
      expect(alertStateLabel("open")).toBe("Open");
      expect(alertStateLabel("in_progress")).toBe("In Progress");
      expect(alertStateLabel("solved")).toBe("Solved");
      expect(alertStateLabel("cancelled")).toBe("Cancelled");
    });
  });

  describe("testTypeLabel", () => {
    it("maps all test types to translation keys", () => {
      expect(testTypeLabel("pass_fail")).toBe("Pass fail");
      expect(testTypeLabel("measure")).toBe("Measure");
      expect(testTypeLabel("instruction")).toBe("Instruction");
    });
  });

  describe("severityLabel", () => {
    it("maps severity levels to translation keys", () => {
      expect(severityLabel("low")).toBe("Low");
      expect(severityLabel("medium")).toBe("Medium");
      expect(severityLabel("high")).toBe("High");
      expect(severityLabel("critical")).toBe("Critical");
    });
  });

  describe("checkResultTone", () => {
    it("returns correct tone for each result", () => {
      expect(checkResultTone("pending")).toBe("info");
      expect(checkResultTone("pass")).toBe("success");
      expect(checkResultTone("fail")).toBe("danger");
    });
  });

  describe("alertStateTone", () => {
    it("returns correct tone for each state", () => {
      expect(alertStateTone("open")).toBe("danger");
      expect(alertStateTone("in_progress")).toBe("warning");
      expect(alertStateTone("solved")).toBe("success");
      expect(alertStateTone("cancelled")).toBe("neutral");
    });
  });

  describe("canRecordResult", () => {
    it("allows result only for pending checks", () => {
      expect(canRecordResult({ result: "pending" } as QualityCheck)).toBe(true);
      expect(canRecordResult({ result: "pass" } as QualityCheck)).toBe(false);
      expect(canRecordResult({ result: "fail" } as QualityCheck)).toBe(false);
    });
  });

  describe("canTransitionAlert", () => {
    it("allows open → in_progress", () => {
      expect(canTransitionAlert("open", "in_progress")).toBe(true);
    });

    it("allows open → cancelled", () => {
      expect(canTransitionAlert("open", "cancelled")).toBe(true);
    });

    it("allows in_progress → solved", () => {
      expect(canTransitionAlert("in_progress", "solved")).toBe(true);
    });

    it("allows in_progress → cancelled", () => {
      expect(canTransitionAlert("in_progress", "cancelled")).toBe(true);
    });

    it("rejects invalid transitions", () => {
      expect(canTransitionAlert("open", "solved")).toBe(false);
      expect(canTransitionAlert("solved", "open")).toBe(false);
      expect(canTransitionAlert("cancelled", "open")).toBe(false);
      expect(canTransitionAlert("open", "open")).toBe(false);
    });
  });

  describe("alertStateSteps", () => {
    it("has correct order", () => {
      expect(alertStateSteps).toEqual(["open", "in_progress", "solved"]);
    });
  });

  describe("alertStateIndex", () => {
    it("returns correct index for each state", () => {
      expect(alertStateIndex("open")).toBe(0);
      expect(alertStateIndex("in_progress")).toBe(1);
      expect(alertStateIndex("solved")).toBe(2);
      expect(alertStateIndex("cancelled")).toBe(-1);
    });
  });

  describe("checksWithFailedCount", () => {
    it("counts failed checks", () => {
      expect(
        checksWithFailedCount([
          { result: "pass" },
          { result: "fail" },
          { result: "fail" },
          { result: "pending" },
        ] as QualityCheck[]),
      ).toBe(2);
    });

    it("returns 0 for empty array", () => {
      expect(checksWithFailedCount([])).toBe(0);
    });
  });

  describe("checksPendingCount", () => {
    it("counts pending checks", () => {
      expect(
        checksPendingCount([
          { result: "pending" },
          { result: "pass" },
          { result: "pending" },
        ] as QualityCheck[]),
      ).toBe(2);
    });

    it("returns 0 for empty array", () => {
      expect(checksPendingCount([])).toBe(0);
    });
  });

  describe("checksPassRate", () => {
    it("calculates pass rate from resolved checks", () => {
      expect(
        checksPassRate([
          { result: "pass" },
          { result: "pass" },
          { result: "fail" },
          { result: "pending" },
        ] as QualityCheck[]),
      ).toBe(67);
    });

    it("returns 0 for no resolved checks", () => {
      expect(checksPassRate([{ result: "pending" }] as QualityCheck[])).toBe(0);
    });

    it("returns 100 for all pass", () => {
      expect(checksPassRate([{ result: "pass" }, { result: "pass" }] as QualityCheck[])).toBe(100);
    });
  });

  describe("isMeasureType", () => {
    it("returns true for measure type", () => {
      expect(isMeasureType({ testType: "measure" } as QualityPoint)).toBe(true);
    });

    it("returns false for other types", () => {
      expect(isMeasureType({ testType: "pass_fail" } as QualityPoint)).toBe(false);
    });
  });

  describe("isPassFailType", () => {
    it("returns true for pass_fail type", () => {
      expect(isPassFailType({ testType: "pass_fail" } as QualityPoint)).toBe(true);
    });

    it("returns false for other types", () => {
      expect(isPassFailType({ testType: "measure" } as QualityPoint)).toBe(false);
    });
  });

  describe("normRangeLabel", () => {
    it("shows both bounds", () => {
      expect(normRangeLabel({ normMin: 10, normMax: 20 } as QualityPoint)).toBe("10 – 20");
    });

    it("shows min only", () => {
      expect(normRangeLabel({ normMin: 5, normMax: null } as QualityPoint)).toBe("≥ 5");
    });

    it("shows max only", () => {
      expect(normRangeLabel({ normMin: null, normMax: 100 } as QualityPoint)).toBe("≤ 100");
    });

    it("shows dash for no norms", () => {
      expect(normRangeLabel({ normMin: null, normMax: null } as QualityPoint)).toBe("—");
    });
  });

  describe("alertTraceability", () => {
    it("builds traceability string from available fields", () => {
      expect(
        alertTraceability({
          checkId: 42,
          itemId: 10,
          batchId: 5,
        } as QualityAlert),
      ).toBe("Check #42 · Item #10 · Batch #5");
    });

    it("shows dash when no fields present", () => {
      expect(
        alertTraceability({
          checkId: null,
          itemId: null,
          batchId: null,
        } as QualityAlert),
      ).toBe("—");
    });
  });
});
