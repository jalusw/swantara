import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityCheckDetail } from "./_components/quality-check-detail-section";

export default async function OrgQualityCheckDetailPage({
  params,
}: {
  params: Promise<{ checkId: string }>;
}) {
  const { checkId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/quality/checks"}>{"Back to checks"}</BackLink>
      <QualityCheckDetail orgId={id} checkId={checkId} />
    </div>
  );
}
