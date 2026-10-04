import { describe, expect, it } from "vitest";
import type { ProjectMilestone, ProjectSummary, ProjectTask } from "@/lib/services/swantara";
import {
  billingTypeLabel,
  canCancelProject,
  canCloseProject,
  canOpenProject,
  marginPercent,
  milestoneProgress,
  projectStateIndex,
  projectStateLabel,
  projectStateTone,
  taskStageLabel,
  taskStageTone,
  tasksByStage,
  totalEffectiveHours,
  totalPlannedHours,
  utilizationColor,
  utilizationPercent,
} from "../project-utils";

describe("project-utils", () => {
  describe("projectStateLabel", () => {
    it("maps all states to translation keys", () => {
      expect(projectStateLabel("draft")).toBe("Draft");
      expect(projectStateLabel("open")).toBe("Buka");
      expect(projectStateLabel("closed")).toBe("Ditutup");
      expect(projectStateLabel("cancelled")).toBe("Dibatalkan");
    });
  });

  describe("billingTypeLabel", () => {
    it("maps all billing types to translation keys", () => {
      expect(billingTypeLabel("fixed")).toBe("Tetap");
      expect(billingTypeLabel("time_material")).toBe("Time & Material");
      expect(billingTypeLabel("milestone")).toBe("Tonggak");
    });
  });

  describe("canOpenProject", () => {
    it("allows open only in draft", () => {
      expect(canOpenProject("draft")).toBe(true);
      expect(canOpenProject("open")).toBe(false);
      expect(canOpenProject("closed")).toBe(false);
      expect(canOpenProject("cancelled")).toBe(false);
    });
  });

  describe("canCloseProject", () => {
    it("allows close only in open", () => {
      expect(canCloseProject("open")).toBe(true);
      expect(canCloseProject("draft")).toBe(false);
      expect(canCloseProject("closed")).toBe(false);
    });
  });

  describe("canCancelProject", () => {
    it("allows cancel in draft and open", () => {
      expect(canCancelProject("draft")).toBe(true);
      expect(canCancelProject("open")).toBe(true);
      expect(canCancelProject("closed")).toBe(false);
      expect(canCancelProject("cancelled")).toBe(false);
    });
  });

  describe("projectStateTone", () => {
    it("returns correct tone for each state", () => {
      expect(projectStateTone("draft")).toBe("neutral");
      expect(projectStateTone("open")).toBe("info");
      expect(projectStateTone("closed")).toBe("success");
      expect(projectStateTone("cancelled")).toBe("danger");
    });
  });

  describe("projectStateIndex", () => {
    it("returns correct index for each state", () => {
      expect(projectStateIndex("draft")).toBe(0);
      expect(projectStateIndex("open")).toBe(1);
      expect(projectStateIndex("closed")).toBe(2);
      expect(projectStateIndex("cancelled")).toBe(-1);
    });
  });

  describe("taskStageLabel", () => {
    it("maps all stages to translation keys", () => {
      expect(taskStageLabel("backlog")).toBe("Backlog");
      expect(taskStageLabel("todo")).toBe("To Do");
      expect(taskStageLabel("in_progress")).toBe("Berjalan");
      expect(taskStageLabel("done")).toBe("Selesai");
    });
  });

  describe("taskStageTone", () => {
    it("returns correct tone for each stage", () => {
      expect(taskStageTone("backlog")).toBe("neutral");
      expect(taskStageTone("todo")).toBe("info");
      expect(taskStageTone("in_progress")).toBe("warning");
      expect(taskStageTone("done")).toBe("success");
    });
  });

  describe("utilizationPercent", () => {
    it("calculates utilization from summary", () => {
      expect(
        utilizationPercent({
          plannedHours: 100,
          effectiveHours: 50,
        } as ProjectSummary),
      ).toBe(50);
      expect(
        utilizationPercent({
          plannedHours: 0,
          effectiveHours: 0,
        } as ProjectSummary),
      ).toBe(0);
      expect(
        utilizationPercent({
          plannedHours: 80,
          effectiveHours: 100,
        } as ProjectSummary),
      ).toBe(100);
    });
  });

  describe("utilizationColor", () => {
    it("returns correct color for utilization percentage", () => {
      expect(utilizationColor(100)).toBe("text-success");
      expect(utilizationColor(75)).toBe("text-warning");
      expect(utilizationColor(50)).toBe("text-muted-foreground");
    });
  });

  describe("marginPercent", () => {
    it("calculates margin from summary", () => {
      expect(
        marginPercent({
          billableAmount: 1000,
          costAmount: 600,
        } as ProjectSummary),
      ).toBe(40);
      expect(
        marginPercent({
          billableAmount: 0,
          costAmount: 0,
        } as ProjectSummary),
      ).toBe(0);
    });
  });

  describe("milestoneProgress", () => {
    it("calculates milestone progress", () => {
      expect(
        milestoneProgress([{ reached: true }, { reached: false }, { reached: true }] satisfies Pick<
          ProjectMilestone,
          "reached"
        >[]),
      ).toBe(67);
      expect(milestoneProgress([])).toBe(0);
      expect(
        milestoneProgress([{ reached: true }] satisfies Pick<ProjectMilestone, "reached">[]),
      ).toBe(100);
    });
  });

  describe("tasksByStage", () => {
    it("counts tasks by stage", () => {
      const tasks: Pick<ProjectTask, "stage">[] = [
        { stage: "backlog" },
        { stage: "todo" },
        { stage: "todo" },
        { stage: "in_progress" },
        { stage: "done" },
        { stage: "done" },
        { stage: "done" },
      ];
      expect(tasksByStage(tasks)).toEqual({
        backlog: 1,
        todo: 2,
        in_progress: 1,
        done: 3,
      });
    });

    it("returns zeros for empty array", () => {
      expect(tasksByStage([])).toEqual({
        backlog: 0,
        todo: 0,
        in_progress: 0,
        done: 0,
      });
    });
  });

  describe("totalPlannedHours", () => {
    it("sums planned hours", () => {
      expect(
        totalPlannedHours([
          { plannedHours: 10 },
          { plannedHours: 20 },
          { plannedHours: 5 },
        ] satisfies Pick<ProjectTask, "plannedHours">[]),
      ).toBe(35);
    });

    it("returns 0 for empty array", () => {
      expect(totalPlannedHours([])).toBe(0);
    });
  });

  describe("totalEffectiveHours", () => {
    it("sums effective hours", () => {
      expect(
        totalEffectiveHours([{ effectiveHours: 8 }, { effectiveHours: 12 }] satisfies Pick<
          ProjectTask,
          "effectiveHours"
        >[]),
      ).toBe(20);
    });

    it("returns 0 for empty array", () => {
      expect(totalEffectiveHours([])).toBe(0);
    });
  });
});
