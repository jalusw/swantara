import type {
  Project,
  ProjectMilestone,
  ProjectSummary,
  ProjectTask,
} from "@/lib/services/swantara";

export type ProjectState = Project["state"];

export function projectStateLabel(state: ProjectState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "open":
      return "Open";
    case "closed":
      return "Closed";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}

export function billingTypeLabel(type: Project["billingType"]): string {
  switch (type) {
    case "fixed":
      return "Fixed";
    case "time_material":
      return "Time & Material";
    case "milestone":
      return "Milestone";
    default:
      return type;
  }
}

export function canOpenProject(state: ProjectState): boolean {
  return state === "draft";
}

export function canCloseProject(state: ProjectState): boolean {
  return state === "open";
}

export function canCancelProject(state: ProjectState): boolean {
  return state === "draft" || state === "open";
}

export function projectStateTone(
  state: ProjectState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "open":
      return "info";
    case "closed":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function projectStateIndex(state: ProjectState): number {
  const steps: ProjectState[] = ["draft", "open", "closed"];
  return steps.indexOf(state);
}

export type TaskStage = ProjectTask["stage"];

export function taskStageLabel(stage: TaskStage): string {
  switch (stage) {
    case "backlog":
      return "Backlog";
    case "todo":
      return "To Do";
    case "in_progress":
      return "In Progress";
    case "done":
      return "Done";
    default:
      return stage;
  }
}

export function taskPriorityLabel(priority: number): string {
  switch (priority) {
    case 0:
      return "None";
    case 1:
      return "Low";
    case 2:
      return "Medium";
    case 3:
      return "High";
    case 4:
      return "Urgent";
    default:
      return String(priority);
  }
}

export function taskStageTone(stage: TaskStage): "neutral" | "success" | "warning" | "info" {
  switch (stage) {
    case "backlog":
      return "neutral";
    case "todo":
      return "info";
    case "in_progress":
      return "warning";
    case "done":
      return "success";
    default:
      return "neutral";
  }
}

export function utilizationPercent(summary: ProjectSummary): number {
  if (summary.plannedHours === 0) return 0;
  return Math.min(100, Math.round((summary.effectiveHours / summary.plannedHours) * 100));
}

export function utilizationColor(pct: number): string {
  if (pct >= 100) return "text-success";
  if (pct >= 75) return "text-warning";
  return "text-muted-foreground";
}

export function marginPercent(summary: ProjectSummary): number {
  if (summary.billableAmount === 0) return 0;
  return Math.round(((summary.billableAmount - summary.costAmount) / summary.billableAmount) * 100);
}

export function milestoneProgress(milestones: Pick<ProjectMilestone, "reached">[]): number {
  if (milestones.length === 0) return 0;
  const reached = milestones.filter((m) => m.reached).length;
  return Math.round((reached / milestones.length) * 100);
}

export function tasksByStage(tasks: Pick<ProjectTask, "stage">[]): Record<TaskStage, number> {
  const counts: Record<TaskStage, number> = {
    backlog: 0,
    todo: 0,
    in_progress: 0,
    done: 0,
  };
  for (const task of tasks) {
    counts[task.stage]++;
  }
  return counts;
}

export function totalPlannedHours(tasks: Pick<ProjectTask, "plannedHours">[]): number {
  return tasks.reduce((sum, task) => sum + task.plannedHours, 0);
}

export function totalEffectiveHours(tasks: Pick<ProjectTask, "effectiveHours">[]): number {
  return tasks.reduce((sum, task) => sum + task.effectiveHours, 0);
}
