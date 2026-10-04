"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Avatar, AvatarFallback } from "@/components/avatar";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Employee } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";
import { EmployeeFormDialog } from "./employee-form-dialog";

export type EmployeeRow = {
  id: string;
  name: string;
  email: string;
  role: string;
  department: string;
  status: "active" | "inactive";
  joined: string;
  location: string;
};

function toEmployeeRow(employee: Employee): EmployeeRow {
  return {
    id: String(employee.id),
    name: employee.employeeNumber,
    email: "",
    role: employee.employmentType,
    department: "",
    status: employee.active ? "active" : "inactive",
    joined: employee.hireDate ?? "",
    location: employee.workLocation ?? "",
  };
}

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2);
}

export function EmployeesTable({ orgId }: { orgId: string }) {
  const t = useTranslations("Employees");
  const tCommon = useTranslations("Common");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);

  const query = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );

  const employees = (query.data?.employees ?? []).map(toEmployeeRow);

  function handleSave() {
    setDialogOpen(false);
    setEditingEmployee(null);
    void query.refetch();
    toast.success(t("saved"));
  }

  function handleEdit(employee: EmployeeRow) {
    const raw = (query.data?.employees ?? []).find((e) => String(e.id) === employee.id);
    if (raw) {
      setEditingEmployee(raw);
      setDialogOpen(true);
    }
  }

  function handleDelete(employee: EmployeeRow) {
    void getSwantaraService()
      .employees.delete(Number(orgId), Number(employee.id))
      .then(() => {
        toast.success(t("deleted"));
        void query.refetch();
      });
  }

  function statusLabel(status: EmployeeRow["status"]): string {
    try {
      return (t as unknown as (k: string) => string)(`status.${status}`);
    } catch {
      return humanizeKey(String(status));
    }
  }

  const columns: ColumnDef<EmployeeRow>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => (
        <a
          href={`/employees/${row.original.id}`}
          className="flex items-center gap-3 rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <Avatar size="sm">
            <AvatarFallback>{initials(row.original.name)}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span>{row.original.name}</span>
            <span className="text-xs text-muted-foreground">{row.original.email}</span>
          </div>
        </a>
      ),
    },
    {
      accessorKey: "role",
      header: t("tableRole"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.role || "—"}</span>,
    },
    {
      accessorKey: "department",
      header: t("tableDepartment"),
      cell: ({ row }) => <Badge variant="secondary">{row.original.department || "—"}</Badge>,
    },
    {
      accessorKey: "status",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const toneMap: Record<EmployeeRow["status"], "success" | "danger"> = {
          active: "success",
          inactive: "danger",
        };
        return (
          <StateBadge
            tone={toneMap[row.original.status]}
            label={statusLabel(row.original.status)}
          />
        );
      },
    },
    {
      accessorKey: "location",
      header: t("tableLocation"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.location || "—"}</span>
      ),
    },
    {
      accessorKey: "joined",
      header: t("tableJoined"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.joined || "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteEmployeeTitle")}
          confirmDescription={t("deleteEmployeeDescription")}
          onEdit={() => handleEdit(row.original)}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={employees}
        getRowId={(row) => row.id}
        searchKeys={["name", "email", "role", "department"]}
        statusKey="status"
        statusOptions={[
          { value: "active", label: t("statusActive") },
          { value: "inactive", label: t("statusInactive") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("filterByStatus")}
        allLabel={t("allStatuses")}
        ariaLabel={t("allEmployees")}
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
          <Button
            size="sm"
            onClick={() => {
              setEditingEmployee(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{t("addEmployee")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <EmployeeFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editingEmployee}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
