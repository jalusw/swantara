import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { EmployeesStats } from "./_components/employees-stats-section";
import { EmployeesTable } from "./_components/employees-table-section";

export default async function OrgEmployeesPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Employees");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("subtitle")} />

      <EmployeesStats />

      <Card>
        <CardHeader className="border-b">
          <CardTitle>{t("allEmployees")}</CardTitle>
        </CardHeader>
        <CardContent>
          <EmployeesTable orgId={id} />
        </CardContent>
      </Card>
    </div>
  );
}
