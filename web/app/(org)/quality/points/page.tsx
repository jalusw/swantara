import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityPointsSection } from "./_components/quality-points-section";

export default async function OrgQualityPointsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Quality Points"}
        description={"Configure quality inspection points for products and operations."}
      />
      <QualityPointsSection orgId={id} />
    </div>
  );
}
