import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CommissionPlanDetail } from "./_components/commission-plan-detail-section";

export default async function CommissionPlanDetailPage({
  params,
}: {
  params: Promise<{ planId: string }>;
}) {
  const t = await getTranslations("Commissions");
  const { planId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/commission-plans"}>{t("backToPlans")}</BackLink>
      <CommissionPlanDetail orgId={id} planId={planId} />
    </div>
  );
}
