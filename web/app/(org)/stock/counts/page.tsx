import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";

import { CountsSection } from "./_components/counts-section";

export default async function OrgCountsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Stock",
  );
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("countsTitle")} description={t("countsDescription")} />
      <CountsSection />
    </div>
  );
}
