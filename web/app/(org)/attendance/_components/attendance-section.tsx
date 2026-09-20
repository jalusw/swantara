"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { LogIn } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Attendance, Employee } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDateTime, formatHours } from "@/lib/utils";
import { isCheckedIn } from "./attendance-utils";
import { CheckInDialog } from "./check-in-dialog";

export function AttendanceSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ attendances: Attendance[] }, Record<string, never>>(
    "attendances",
    (organizationId) => getSwantaraService().attendances.list(organizationId),
  );

  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );

  const attendances = query.data?.attendances ?? [];
  const employees = employeesQuery.data?.employees ?? [];

  const employeeNameMap = new Map(employees.map((emp) => [emp.id, emp.employeeNumber]));

  function handleCheckIn() {
    setDialogOpen(false);
    void query.refetch();
  }

  const checkOutMutation = useMutation({
    mutationFn: (attendance: Attendance) =>
      getSwantaraService().attendances.checkOut(Number(orgId), attendance.id, {
        checkOut: new Date().toISOString(),
      }),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["attendances"] }),
    onError: () => toast.error("Could not disable the organization."),
  });

  function handleCheckOut(attendance: Attendance) {
    checkOutMutation.mutate(attendance);
  }

  const columns: ColumnDef<Attendance>[] = [
    {
      accessorKey: "employeeId",
      header: "Employee",
      cell: ({ row }) => (
        <span className="">{employeeNameMap.get(row.original.employeeId) ?? "—"}</span>
      ),
    },
    {
      accessorKey: "checkIn",
      header: "Check in",
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatDateTime(row.original.checkIn, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "checkOut",
      header: "Check out",
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatDateTime(row.original.checkOut, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "workedHours",
      header: "Worked Hours",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatHours(row.original.workedHours)}
        </span>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: ({ row }) => (
        <ActiveBadge active={isCheckedIn(row.original)}>
          {isCheckedIn(row.original) ? "Checked In" : "Completed"}
        </ActiveBadge>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        if (!isCheckedIn(row.original)) return null;
        return (
          <Button
            variant="ghost"
            size="sm"
            disabled={checkOutMutation.isPending}
            onClick={() => handleCheckOut(row.original)}
          >
            {"Check out"}
          </Button>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={attendances}
        getRowId={(row) => String(row.id)}
        searchKeys={["employeeId"]}
        searchPlaceholder={"Search attendance…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"Attendance"}
        emptyTitle={"No attendance records found"}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <LogIn />
            <span>{"Check in"}</span>
          </Button>
        }
      />

      <CheckInDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        orgId={orgId}
        employees={employees}
        onSave={handleCheckIn}
      />
    </div>
  );
}
