import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CommissionPlanDetail } from "./_components/commission-plan-detail-section";

export default async function CommissionPlanDetailPage({
  params,
}: {
  params: Promise<{ planId: string }>;
}) {
  const { planId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/commission-plans"}>{"Back to plans"}</BackLink>
      <CommissionPlanDetail orgId={id} planId={planId} />
    </div>
  );
}
