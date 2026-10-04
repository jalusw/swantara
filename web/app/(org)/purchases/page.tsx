import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PurchaseOrdersSection } from "./orders/_components/purchase-orders-section";

export default async function OrgPurchaseOrdersPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("ordersTitle")} description={t("ordersSubtitle")} />
      <PurchaseOrdersSection orgId={id} />
    </div>
  );
}
