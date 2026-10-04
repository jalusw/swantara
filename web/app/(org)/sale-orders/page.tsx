import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SaleOrdersSection } from "./_components/sale-orders-section";

export default async function OrgSaleOrdersPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Sales");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("ordersTitle")} description={t("ordersSubtitle")} />
      <SaleOrdersSection orgId={id} />
    </div>
  );
}
