import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DeferralDetail } from "./_components/deferral-detail-section";

export default async function DeferralDetailPage({
  params,
}: {
  params: Promise<{ deferralId: string }>;
}) {
  const { deferralId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/deferrals"}>{"Back to deferrals"}</BackLink>
      <DeferralDetail orgId={id} deferralId={deferralId} />
    </div>
  );
}
