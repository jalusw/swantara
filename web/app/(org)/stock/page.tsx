import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { StockOverviewSection } from "./_components/stock-overview-section";

export default async function OrgStockPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Stock",
  );
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("overviewTitle")} description={t("overviewDescription")} />
      <StockOverviewSection orgId={id} />
    </div>
  );
}
