import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { ReportsHubSection } from "./_components/reports-hub-section";

export default async function OrgReportsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Reports",
  );
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("reportsTitle")} description={t("reportsDescription")} />
      <ReportsHubSection />
    </div>
  );
}
