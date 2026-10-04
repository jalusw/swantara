"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Employee, LeaveRequest, LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import {
  canApprove,
  canRefuse,
  canSubmit,
  leaveRequestStateTone,
} from "../types/_components/leave-type-utils";
import { LeaveRequestFormDialog } from "./leave-request-form-dialog";

export function LeaveRequestsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Leave");
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);

  const leaveRequestsQuery = useOrgListQuery<
    { leaveRequests: LeaveRequest[] },
    Record<string, never>
  >("leaveRequests", (organizationId) => getSwantaraService().leaveRequests.list(organizationId));
  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const leaveTypesQuery = useOrgListQuery<{ leaveTypes: LeaveType[] }, Record<string, never>>(
    "leaveTypes",
    (organizationId) => getSwantaraService().leaveTypes.list(organizationId),
  );

  const leaveRequests = leaveRequestsQuery.data?.leaveRequests ?? [];
  const employees = employeesQuery.data?.employees ?? [];
  const contacts = contactsQuery.data?.contacts ?? [];
  const leaveTypes = leaveTypesQuery.data?.leaveTypes ?? [];

  const contactMap = useMemo(
    () => new Map(contacts.map((p) => [p.id, p.displayName || p.name])),
    [contacts],
  );
  const employeeNameMap = useMemo(
    () => new Map(employees.map((e) => [e.id, contactMap.get(e.contactId) ?? `#${e.id}`])),
    [employees, contactMap],
  );
  const leaveTypeMap = useMemo(
    () => new Map(leaveTypes.map((lt) => [lt.id, lt.name])),
    [leaveTypes],
  );

  function handleSave() {
    setDialogOpen(false);
    void leaveRequestsQuery.refetch();
  }

  const actionMutation = useMutation({
    mutationFn: ({ action, id }: { action: "submit" | "approve" | "refuse"; id: number }) => {
      const service = getSwantaraService().leaveRequests;
      return action === "submit"
        ? service.submit(Number(orgId), id)
        : action === "approve"
          ? service.approve(Number(orgId), id)
          : service.refuse(Number(orgId), id);
    },
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["leaveRequests"] }),
  });

  function handleAction(action: "submit" | "approve" | "refuse", id: number) {
    actionMutation.mutate({ action, id });
  }

  function stateLabel(state: LeaveRequest["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`state.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<LeaveRequest>[] = [
    {
      accessorKey: "employeeId",
      header: t("fieldEmployee"),
      cell: ({ row }) => (
        <span className="">
          {employeeNameMap.get(row.original.employeeId) ?? `#${row.original.employeeId}`}
        </span>
      ),
    },
    {
      accessorKey: "leaveTypeId",
      header: t("fieldLeaveType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {leaveTypeMap.get(row.original.leaveTypeId) ?? `#${row.original.leaveTypeId}`}
        </span>
      ),
    },
    {
      accessorKey: "dateFrom",
      header: t("tablePeriod"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {formatDate(row.original.dateFrom)} – {formatDate(row.original.dateTo)}
        </span>
      ),
    },
    {
      accessorKey: "days",
      header: t("fieldDays"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatNumber(row.original.days)}</span>,
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const tone = leaveRequestStateTone(row.original.state);
        return (
          <Badge variant="outline" className={tone}>
            {stateLabel(row.original.state)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const state = row.original.state;
        const id = row.original.id;
        return (
          <div className="flex items-center gap-1">
            {canSubmit(state) ? (
              <Button
                size="sm"
                variant="outline"
                disabled={actionMutation.isPending}
                onClick={() => handleAction("submit", id)}
              >
                {t("actionSubmit")}
              </Button>
            ) : null}
            {canApprove(state) ? (
              <Button
                size="sm"
                disabled={actionMutation.isPending}
                onClick={() => handleAction("approve", id)}
              >
                {t("actionApprove")}
              </Button>
            ) : null}
            {canRefuse(state) ? (
              <Button
                size="sm"
                variant="destructive"
                disabled={actionMutation.isPending}
                onClick={() => handleAction("refuse", id)}
              >
                {t("actionRefuse")}
              </Button>
            ) : null}
          </div>
        );
      },
    },
  ];

  const isLoading =
    leaveRequestsQuery.isLoading ||
    employeesQuery.isLoading ||
    contactsQuery.isLoading ||
    leaveTypesQuery.isLoading;
  const error = leaveRequestsQuery.isError
    ? leaveRequestsQuery.error
    : employeesQuery.isError
      ? employeesQuery.error
      : contactsQuery.isError
        ? contactsQuery.error
        : leaveTypesQuery.isError
          ? leaveTypesQuery.error
          : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={leaveRequests}
        getRowId={(row) => String(row.id)}
        searchKeys={["employeeId"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: t("stateDraft") },
          { value: "submitted", label: t("stateSubmitted") },
          { value: "approved", label: t("stateApproved") },
          { value: "refused", label: t("stateRefused") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allLabel")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyRequests")}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void leaveRequestsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("addRequest")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <LeaveRequestFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
