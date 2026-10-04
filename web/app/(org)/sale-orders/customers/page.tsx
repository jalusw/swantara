import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CustomersAcquisition } from "./_components/customers-acquisition-section";
import { CustomersStats } from "./_components/customers-stats-section";
import { CustomersTable } from "./_components/customers-table-section";

export default async function OrgCustomersPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Sales");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("customersTitle")} description={t("customersSubtitle")} />

      <section aria-label={t("customersSummary")}>
        <CustomersStats />
      </section>

      <section aria-label={t("allCustomers")}>
        <Card>
          <CardHeader className="border-b">
            <CardTitle>{t("allCustomers")}</CardTitle>
            <CardDescription>{t("customersDescription")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <CustomersTable orgId={id} />
          </CardContent>
        </Card>
      </section>

      <section aria-label={t("customerAcquisition")}>
        <CustomersAcquisition />
      </section>
    </div>
  );
}
