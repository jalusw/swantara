import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CommissionPlansSection } from "./_components/commission-plans-section";

export default async function CommissionPlansPage() {
  const t = await getTranslations("Commissions");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("plansTitle")} description={t("plansDescription")} />
      <CommissionPlansSection orgId={id} />
    </div>
  );
}
