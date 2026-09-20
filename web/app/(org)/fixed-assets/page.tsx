import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { FixedAssetsSection } from "./_components/fixed-assets-section";

export default async function FixedAssetsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Fixed assets"}
        description={"Register assets, generate depreciation schedules and post periods."}
      />
      <FixedAssetsSection orgId={id} />
    </div>
  );
}
