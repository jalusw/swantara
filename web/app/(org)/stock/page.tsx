import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { StockOverviewSection } from "./_components/stock-overview-section";

export default async function OrgStockPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Stock overview"}
        description={"View on-hand quantities and stock movements across locations."}
      />
      <StockOverviewSection orgId={id} />
    </div>
  );
}
