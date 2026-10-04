"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { ActiveBadge } from "@/components/active-badge";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Employee, EmploymentContract } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { EmployeeFormDialog } from "../../_components/employee-form-dialog";

export function EmployeeDetail({ orgId, employeeId }: { orgId: string; employeeId: string }) {
  const t = useTranslations("Employees");
  const tCommon = useTranslations("Common");
  const [editOpen, setEditOpen] = useState(false);

  const employeeQuery = useOrgQuery<{ employee: Employee }>(
    "employee",
    employeeId,
    (organizationId) => getSwantaraService().employees.get(organizationId, Number(employeeId)),
  );

  const contractsQuery = useOrgListQuery<
    { contracts: EmploymentContract[] },
    Record<string, never>
  >("contracts", (organizationId) => getSwantaraService().contracts.list(organizationId));

  const employee = employeeQuery.data?.employee;
  const contracts = (contractsQuery.data?.contracts ?? []).filter(
    (c) => c.employeeId === Number(employeeId),
  );

  if (employeeQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!employee) {
    return <p className="text-sm text-muted-foreground">{t("detailNotFound")}</p>;
  }

  function handleEditSave() {
    setEditOpen(false);
    toast.success(t("saved"));
    void employeeQuery.refetch();
    void contractsQuery.refetch();
  }

  const status = (
    <ActiveBadge active={employee.active}>
      {employee.active ? t("statusActive") : t("statusInactive")}
    </ActiveBadge>
  );

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: t("title"), href: "/employees" },
          { label: employee.employeeNumber },
        ]}
        title={employee.employeeNumber}
        status={status}
        actions={
          <Button size="sm" variant="outline" onClick={() => setEditOpen(true)}>
            {tCommon("edit")}
          </Button>
        }
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: <EmployeeOverview employee={employee} />,
          },
          {
            id: "contracts",
            label: t("tabContracts"),
            content: (
              <EmployeeContracts
                contracts={contracts}
                status={
                  contractsQuery.isLoading
                    ? { type: "loading" }
                    : contractsQuery.isError
                      ? {
                          type: "error",
                          message: contractsQuery.error.message,
                          onRetry: () => void contractsQuery.refetch(),
                        }
                      : undefined
                }
              />
            ),
          },
        ]}
      />
      {editOpen ? (
        <EmployeeFormDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          orgId={orgId}
          initial={employee}
          onSave={handleEditSave}
        />
      ) : null}
    </>
  );
}

function EmployeeOverview({ employee }: { employee: Employee }) {
  const t = useTranslations("Employees");
  const identity = [
    { label: t("fieldEmployeeNumber"), value: employee.employeeNumber },
    {
      label: t("fieldEmploymentType"),
      value: String(employee.employmentType),
    },
    { label: t("fieldHireDate"), value: employee.hireDate ?? "—" },
    { label: t("fieldTerminationDate"), value: employee.terminationDate ?? "—" },
  ];

  const employment = [
    {
      label: t("fieldDepartment"),
      value: employee.departmentId != null ? String(employee.departmentId) : "—",
    },
    {
      label: t("fieldJobPosition"),
      value: employee.jobPositionId != null ? String(employee.jobPositionId) : "—",
    },
    { label: t("fieldWorkLocation"), value: employee.workLocation ?? "—" },
  ];

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardContent>
          <Section title={t("sectionIdentity")} items={identity} />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Section title={t("formEmployment")} items={employment} />
        </CardContent>
      </Card>
    </div>
  );
}

function EmployeeContracts({
  contracts,
  status,
}: {
  contracts: EmploymentContract[];
  status?: {
    type: "loading" | "error";
    message?: string;
    onRetry?: () => void;
  };
}) {
  const t = useTranslations("Employees");
  const tCommon = useTranslations("Common");
  if (status?.type === "loading") {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (status?.type === "error") {
    return (
      <div className="flex flex-col items-center gap-2 py-8">
        <p className="text-sm text-muted-foreground">{status.message}</p>
        <Button variant="outline" size="sm" onClick={status.onRetry}>
          {tCommon("retry")}
        </Button>
      </div>
    );
  }

  if (contracts.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noContracts")}</p>;
  }

  return (
    <div className="space-y-3">
      {contracts.map((contract) => (
        <Card key={contract.id}>
          <CardContent>
            <dl className="grid gap-3 sm:grid-cols-2">
              <dt className="text-xs text-muted-foreground">{t("contractPeriod")}</dt>
              <dd className="text-sm">
                {contract.dateStart} — {contract.dateEnd ?? "—"}
              </dd>
              <dt className="text-xs text-muted-foreground">{t("fieldWage")}</dt>
              <dd className="text-sm">
                {formatNumber(contract.wage)} {contract.currencyCode}
              </dd>
              <dt className="text-xs text-muted-foreground">{t("fieldWageType")}</dt>
              <dd className="text-sm">{String(contract.wageType)}</dd>
              <dt className="text-xs text-muted-foreground">{t("tableStatus")}</dt>
              <dd className="text-sm">
                <ActiveBadge active={contract.state === "active"}>
                  {humanizeKey(String(contract.state))}
                </ActiveBadge>
              </dd>
            </dl>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function Section({ title, items }: { title: string; items: { label: string; value: string }[] }) {
  return (
    <dl className="grid gap-3 sm:grid-cols-2">
      <p className="text-sm sm:col-span-2">{title}</p>
      {items.map((item) => (
        <div key={item.label} className="flex flex-col gap-0.5">
          <dt className="text-xs text-muted-foreground">{item.label}</dt>
          <dd className="text-sm">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
