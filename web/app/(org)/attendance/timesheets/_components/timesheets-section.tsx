"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Employee, Timesheet } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatHours } from "@/lib/utils";
import { TimesheetFormDialog } from "./timesheet-form-dialog";

function toRow(timesheet: Timesheet, employeeName: Map<number, string>) {
  return {
    id: String(timesheet.id),
    employeeId: timesheet.employeeId,
    employeeName: employeeName.get(timesheet.employeeId) ?? String(timesheet.employeeId),
    date: timesheet.date,
    hours: timesheet.hours,
    description: timesheet.description,
    dimensionId: timesheet.dimensionId,
  };
}

export function TimesheetsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Attendance");
  const tCommon = useTranslations("Common");
  const [dialogOpen, setDialogOpen] = useState(false);

  const timesheetsQuery = useOrgListQuery<{ timesheets: Timesheet[] }, Record<string, never>>(
    "timesheets",
    (organizationId) => getSwantaraService().timesheets.list(organizationId),
  );

  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contactMap = new Map(
    (contactsQuery.data?.contacts ?? []).map((p) => [p.id, p.displayName ?? p.name]),
  );

  const employeeNameMap = new Map(
    (employeesQuery.data?.employees ?? []).map((e) => [
      e.id,
      contactMap.get(e.contactId) ?? e.employeeNumber,
    ]),
  );

  const rows = (timesheetsQuery.data?.timesheets ?? []).map((ts) => toRow(ts, employeeNameMap));

  function handleSave() {
    setDialogOpen(false);
    void timesheetsQuery.refetch();
  }

  function handleDelete(row: ReturnType<typeof toRow>) {
    void getSwantaraService()
      .timesheets.delete(Number(orgId), Number(row.id))
      .then(() => void timesheetsQuery.refetch())
      .catch(() => toast.error(t("saveFailed")));
  }

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    {
      accessorKey: "employeeName",
      header: t("fieldEmployee"),
      cell: ({ row }) => <span className="">{row.original.employeeName}</span>,
    },
    {
      accessorKey: "date",
      header: t("fieldDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{formatDate(row.original.date)}</span>
      ),
    },
    {
      accessorKey: "hours",
      header: t("fieldHours"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatHours(row.original.hours)}</span>,
    },
    {
      accessorKey: "description",
      header: t("fieldDescription"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.description || "—"}</span>
      ),
    },
    {
      accessorKey: "dimensionId",
      header: t("fieldDimension"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.dimensionId ?? "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteTimesheetTitle")}
          confirmDescription={t("deleteTimesheetDescription")}
          onEdit={() => {}}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  const isLoading =
    timesheetsQuery.isLoading || employeesQuery.isLoading || contactsQuery.isLoading;
  const error = timesheetsQuery.isError
    ? timesheetsQuery.error
    : employeesQuery.isError
      ? employeesQuery.error
      : contactsQuery.isError
        ? contactsQuery.error
        : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={rows}
        getRowId={(row) => row.id}
        searchKeys={["employeeName", "description"]}
        searchPlaceholder={t("searchTimesheetsPlaceholder")}
        ariaLabel={t("timesheetsTitle")}
        emptyTitle={t("emptyTimesheets")}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => {
                    void timesheetsQuery.refetch();
                    void employeesQuery.refetch();
                    void contactsQuery.refetch();
                  },
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("addTimesheetEntry")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <TimesheetFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
