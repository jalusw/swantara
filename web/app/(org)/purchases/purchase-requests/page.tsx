import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PurchaseRequestsSection } from "./_components/purchase-requests-section";

export default async function OrgPurchaseRequestsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("requisitionsTitle")} description={t("requisitionsSubtitle")} />
      <PurchaseRequestsSection orgId={id} />
    </div>
  );
}
