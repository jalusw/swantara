import { getTranslations } from "next-intl/server";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";

import { SuppliersStats } from "./_components/suppliers-stats-section";
import { SuppliersTable } from "./_components/suppliers-table-section";
import { SUPPLIERS_SPEND_TREND } from "./_utils";

export default async function OrgSuppliersPage() {
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("suppliersTitle")} description={t("suppliersSubtitle")} />

      <SuppliersStats />

      <Card>
        <CardHeader>
          <CardTitle>{t("spendByMonth")}</CardTitle>
          <CardDescription>{t("spendByMonthDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={SUPPLIERS_SPEND_TREND}
            ariaLabel={t("spendByMonth")}
            valueFormatter={(value) => `$${value}k`}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="border-b">
          <CardTitle>{t("allSuppliers")}</CardTitle>
          <CardDescription>{t("suppliersDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <SuppliersTable />
        </CardContent>
      </Card>
    </div>
  );
}
