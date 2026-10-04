import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MoSection } from "./_components/production-order-section";

export default async function OrgProductionOrdersPage() {
  const t = await getTranslations("ProductionOrders");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <MoSection orgId={id} />
    </div>
  );
}
