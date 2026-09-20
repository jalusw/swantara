import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityAlertDetail } from "./_components/quality-alert-detail-section";

export default async function OrgQualityAlertDetailPage({
  params,
}: {
  params: Promise<{ alertId: string }>;
}) {
  const { alertId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/quality/alerts"}>{"Back to alerts"}</BackLink>
      <QualityAlertDetail orgId={id} alertId={alertId} />
    </div>
  );
}
