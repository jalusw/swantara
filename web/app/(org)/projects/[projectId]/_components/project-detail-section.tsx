"use client";

import { Pencil, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { RecordLayout } from "@/components/record-layout";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  Journal,
  Project,
  ProjectMilestone,
  ProjectSummary,
  ProjectTask,
  Timesheet,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber, getLocalDateString } from "@/lib/utils";
import { ProjectFormDialog } from "../../_components/project-form-dialog";
import {
  canCancelProject,
  canCloseProject,
  canOpenProject,
  marginPercent,
  utilizationPercent,
} from "../../_components/project-utils";
import { MilestoneFormDialog } from "./milestone-form-dialog";
import { TaskFormDialog } from "./task-form-dialog";

export function ProjectDetail({ orgId, projectId }: { orgId: string; projectId: string }) {
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const dyn = (key: string) => (t as unknown as (k: string) => string)(key);
  const [editOpen, setEditOpen] = useState(false);
  const [taskDialogOpen, setTaskDialogOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<ProjectTask | null>(null);
  const [milestoneDialogOpen, setMilestoneDialogOpen] = useState(false);
  const [editingMilestone, setEditingMilestone] = useState<ProjectMilestone | null>(null);
  const [billOpen, setBillOpen] = useState(false);

  const projectQuery = useOrgQuery<{ project: Project }>("project", projectId, (organizationId) =>
    getSwantaraService().projects.get(organizationId, Number(projectId)),
  );

  const summaryQuery = useOrgQuery<{ summary: ProjectSummary }>(
    "projectSummary",
    projectId,
    (organizationId) => getSwantaraService().projects.summary(organizationId, Number(projectId)),
  );

  const tasksQuery = useOrgListQuery<{ tasks: ProjectTask[] }, Record<string, never>>(
    "projectTasks",
    (organizationId) => getSwantaraService().projects.tasks.list(organizationId, Number(projectId)),
  );

  const milestonesQuery = useOrgListQuery<
    { milestones: ProjectMilestone[] },
    Record<string, never>
  >("projectMilestones", (organizationId) =>
    getSwantaraService().projects.milestones.list(organizationId, Number(projectId)),
  );

  const timesheetsQuery = useOrgListQuery<{ timesheets: Timesheet[] }, Record<string, never>>(
    "timesheets",
    (organizationId) => getSwantaraService().timesheets.list(organizationId),
  );

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const project = projectQuery.data?.project ?? null;
  const summary = summaryQuery.data?.summary ?? null;
  const tasks = tasksQuery.data?.tasks ?? [];
  const milestones = milestonesQuery.data?.milestones ?? [];
  const contactMap = new Map(
    (contactsQuery.data?.contacts ?? []).map((p) => [p.id, p.displayName ?? p.name]),
  );

  const projectTimesheets = (timesheetsQuery.data?.timesheets ?? []).filter(
    (ts) => ts.projectId === Number(projectId),
  );

  const isLoading =
    projectQuery.isLoading ||
    summaryQuery.isLoading ||
    tasksQuery.isLoading ||
    milestonesQuery.isLoading;

  function handleOpen() {
    if (!project) return;
    void getSwantaraService()
      .projects.state(Number(orgId), project.id, { state: "open" })
      .then(() => {
        toast.success(t("projectOpened"));
        void projectQuery.refetch();
      })
      .catch(() => toast.error(t("statusUpdateFailed")));
  }

  function handleClose() {
    if (!project) return;
    void getSwantaraService()
      .projects.state(Number(orgId), project.id, { state: "closed" })
      .then(() => {
        toast.success(t("projectClosed"));
        void projectQuery.refetch();
      })
      .catch(() => toast.error(t("statusUpdateFailed")));
  }

  function handleCancel() {
    if (!project) return;
    void getSwantaraService()
      .projects.state(Number(orgId), project.id, { state: "cancelled" })
      .then(() => {
        toast.success(t("projectCancelled"));
        void projectQuery.refetch();
      })
      .catch(() => toast.error(t("statusUpdateFailed")));
  }

  if (isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!project) {
    return <p className="text-sm text-muted-foreground">{t("projectNotFound")}</p>;
  }

  const utilPct = summary ? utilizationPercent(summary) : 0;
  const margPct = summary ? marginPercent(summary) : 0;

  const overviewContent = (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div>
          <p className="text-sm text-muted-foreground">{t("customer")}</p>
          <p className="">{contactMap.get(project.contactId) ?? "—"}</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{t("manager")}</p>
          <p className="">{project.managerId ? (contactMap.get(project.managerId) ?? "—") : "—"}</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{t("billingType")}</p>
          <Badge variant="secondary">{dyn(`billing_${project.billingType}`)}</Badge>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{t("billableRate")}</p>
          <p className=" tabular-nums">{formatNumber(project.billableRate)}</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{t("startDate")}</p>
          <p className="">{formatDate(project.dateStart, { nullFallback: "—" })}</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{t("endDate")}</p>
          <p className="">{formatDate(project.dateEnd, { nullFallback: "—" })}</p>
        </div>
      </div>
      {summary ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("plannedHours")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatNumber(summary.plannedHours)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("effectiveHours")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatNumber(summary.effectiveHours)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("utilization")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">{utilPct}%</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("marginPct")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">{margPct}%</p>
            </CardContent>
          </Card>
        </div>
      ) : null}
    </div>
  );

  const tasksContent = (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button
          size="sm"
          onClick={() => {
            setEditingTask(null);
            setTaskDialogOpen(true);
          }}
        >
          <Plus />
          <span>{t("newTask")}</span>
        </Button>
      </div>
      {tasks.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noTasks")}</p>
      ) : (
        <div className="flex flex-col gap-2">
          {tasks.map((task) => (
            <div key={task.id} className="flex items-center justify-between rounded-md border p-3">
              <div className="flex flex-col gap-1">
                <p className="">{task.name}</p>
                <div className="flex gap-2 text-sm text-muted-foreground">
                  <Badge variant="secondary">{dyn(`stage_${task.stage}`)}</Badge>
                  <span>
                    {task.effectiveHours}/{task.plannedHours}h
                  </span>
                  {task.deadline ? (
                    <span>
                      {t("due")} {formatDate(task.deadline)}
                    </span>
                  ) : null}
                </div>
              </div>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  setEditingTask(task);
                  setTaskDialogOpen(true);
                }}
              >
                <Pencil className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const milestonesContent = (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button
          size="sm"
          onClick={() => {
            setEditingMilestone(null);
            setMilestoneDialogOpen(true);
          }}
        >
          <Plus />
          <span>{t("newMilestone")}</span>
        </Button>
      </div>
      {milestones.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noMilestones")}</p>
      ) : (
        <div className="flex flex-col gap-2">
          {milestones.map((ms) => (
            <div key={ms.id} className="flex items-center justify-between rounded-md border p-3">
              <div className="flex flex-col gap-1">
                <p className="">{ms.name}</p>
                <div className="flex gap-2 text-sm text-muted-foreground">
                  <Badge variant={ms.reached ? "default" : "secondary"}>
                    {ms.reached ? t("milestoneReached") : t("milestoneOpen")}
                  </Badge>
                  {ms.deadline ? (
                    <span>
                      {t("due")} {formatDate(ms.deadline)}
                    </span>
                  ) : null}
                </div>
              </div>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  void getSwantaraService()
                    .projects.milestones.reached(Number(orgId), project.id, ms.id, {
                      reached: !ms.reached,
                    })
                    .then(() => {
                      toast.success(t("milestoneUpdated"));
                      void milestonesQuery.refetch();
                    });
                }}
              >
                {ms.reached ? t("markUnreached") : t("markReached")}
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const timesheetsContent = (
    <div className="flex flex-col gap-4">
      {projectTimesheets.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noTimesheets")}</p>
      ) : (
        <div className="flex flex-col gap-2">
          {projectTimesheets.map((ts) => (
            <div key={ts.id} className="flex items-center justify-between rounded-md border p-3">
              <div className="flex flex-col gap-1">
                <p className="">{ts.description ?? t("timesheetEntry")}</p>
                <div className="flex gap-2 text-sm text-muted-foreground">
                  <span>{formatDate(ts.date)}</span>
                  <span>{ts.hours}h</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const billingContent =
    project.billingType === "time_material" ? (
      <BillingPanel project={project} summary={summary} />
    ) : (
      <p className="text-sm text-muted-foreground">{t("billingOnlyTimeMaterial")}</p>
    );

  const pnlContent = summary ? (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">{t("revenue")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">
              {formatNumber(summary.billableAmount)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">{t("cost")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{formatNumber(summary.costAmount)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">{t("margin")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">
              {formatNumber(summary.marginAmount)} ({margPct}%)
            </p>
          </CardContent>
        </Card>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">{t("utilization")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{utilPct}%</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">
              {t("plannedVsEffective")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">
              {formatNumber(summary.effectiveHours)} / {formatNumber(summary.plannedHours)}h
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  ) : (
    <p className="text-sm text-muted-foreground">{t("noFinancialData")}</p>
  );

  return (
    <>
      <RecordLayout
        breadcrumbItems={[{ label: t("allProjects"), href: "/projects" }, { label: project.name }]}
        title={project.name}
        status={
          <StateBadge
            value={project.state}
            statuses={{
              draft: { label: dyn("state_draft"), tone: "neutral" },
              open: { label: dyn("state_open"), tone: "info" },
              closed: { label: dyn("state_closed"), tone: "success" },
              cancelled: { label: dyn("state_cancelled"), tone: "danger" },
            }}
          />
        }
        actions={
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil />
              <span>{t("editProject")}</span>
            </Button>
            {canOpenProject(project.state) ? (
              <Button size="sm" onClick={handleOpen}>
                {t("actionOpen")}
              </Button>
            ) : null}
            {canCloseProject(project.state) ? (
              <Button size="sm" onClick={handleClose}>
                {t("actionClose")}
              </Button>
            ) : null}
            {canCancelProject(project.state) ? (
              <Button size="sm" variant="outline" onClick={handleCancel}>
                {tCommon("cancel")}
              </Button>
            ) : null}
          </div>
        }
        tabs={[
          { id: "overview", label: t("tabOverview"), content: overviewContent },
          { id: "tasks", label: t("tabTasks"), content: tasksContent },
          {
            id: "milestones",
            label: t("tabMilestones"),
            content: milestonesContent,
          },
          {
            id: "timesheets",
            label: t("tabTimesheets"),
            content: timesheetsContent,
          },
          { id: "billing", label: t("tabBilling"), content: billingContent },
          { id: "pnl", label: t("tabPnl"), content: pnlContent },
        ]}
      />
      {editOpen ? (
        <ProjectFormDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          orgId={orgId}
          initial={project}
          onSave={() => {
            setEditOpen(false);
            void projectQuery.refetch();
          }}
        />
      ) : null}
      {taskDialogOpen ? (
        <TaskFormDialog
          open={taskDialogOpen}
          onOpenChange={setTaskDialogOpen}
          orgId={orgId}
          projectId={project.id}
          initial={editingTask}
          onSave={() => {
            setTaskDialogOpen(false);
            void tasksQuery.refetch();
          }}
        />
      ) : null}
      {milestoneDialogOpen ? (
        <MilestoneFormDialog
          open={milestoneDialogOpen}
          onOpenChange={setMilestoneDialogOpen}
          orgId={orgId}
          projectId={project.id}
          initial={editingMilestone}
          onSave={() => {
            setMilestoneDialogOpen(false);
            void milestonesQuery.refetch();
          }}
        />
      ) : null}
      {billOpen ? (
        <BillingDialog
          open={billOpen}
          onOpenChange={setBillOpen}
          orgId={orgId}
          project={project}
          summary={summary}
          journalsQuery={journalsQuery}
          onBilled={() => {
            setBillOpen(false);
            void projectQuery.refetch();
            void summaryQuery.refetch();
          }}
        />
      ) : null}
    </>
  );
}

function BillingPanel({ project, summary }: { project: Project; summary: ProjectSummary | null }) {
  const t = useTranslations("Projects");
  return (
    <div className="flex flex-col gap-4">
      {summary ? (
        <div className="grid grid-cols-2 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("unbilledHours")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatNumber(
                  summary.effectiveHours - summary.billedAmount / (project.billableRate || 1),
                )}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{t("unbilledAmount")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatNumber(summary.unbilledAmount)}
              </p>
            </CardContent>
          </Card>
        </div>
      ) : null}
    </div>
  );
}

function BillingDialog({
  open,
  onOpenChange,
  orgId,
  project,
  summary,
  journalsQuery,
  onBilled,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  project: Project;
  summary: ProjectSummary | null;
  journalsQuery: ReturnType<typeof useOrgListQuery<{ journals: Journal[] }, Record<string, never>>>;
  onBilled: () => void;
}) {
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const journalOptions = (
    (journalsQuery.data as { journals: Journal[] } | undefined)?.journals ?? []
  ).map((j: Journal) => ({
    id: String(j.id),
    name: j.name,
  }));

  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());

  function handleBill() {
    if (!journalId || !summary) return;
    void getSwantaraService()
      .projects.bill(Number(orgId), project.id, {
        journalId: Number(journalId),
        date,
        taxIds: [],
      })
      .then((result) => {
        toast.success(
          t("invoiceCreated", {
            name: result.name,
            amount: formatNumber(result.amountTotal),
          }),
        );
        onBilled();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("generateInvoice")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div>
            <p className="text-sm">{t("unbilledAmount")}</p>
            <p className="text-2xl font-bold tabular-nums">
              {summary ? formatNumber(summary.unbilledAmount) : "—"}
            </p>
          </div>
          <div>
            <p className="text-sm">{t("journal")}</p>
            <Select value={journalId} onValueChange={(value) => setJournalId(value ?? "")}>
              <SelectTrigger aria-label={t("journal")}>
                <SelectValue placeholder={t("selectJournal")} />
              </SelectTrigger>
              <SelectContent>
                {journalOptions.map((j) => (
                  <SelectItem key={j.id} value={j.id}>
                    {j.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <p className="text-sm">{t("date")}</p>
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleBill} disabled={!journalId}>
            {t("generateInvoice")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
