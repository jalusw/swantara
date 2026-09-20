import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MaintenancePlanDetail } from "./_components/maintenance-plan-detail-section";

export default async function MaintenancePlanDetailPage({
  params,
}: {
  params: Promise<{ planId: string }>;
}) {
  const { planId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/service-orders/plans"}>{"Back to maintenance plans"}</BackLink>
      <MaintenancePlanDetail orgId={id} planId={planId} />
    </div>
  );
}
