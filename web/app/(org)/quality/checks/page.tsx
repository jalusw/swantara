import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityChecksSection } from "./_components/quality-checks-section";

export default async function OrgQualityChecksPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Quality Checks"}
        description={"Review and record results for quality checks."}
      />
      <QualityChecksSection orgId={id} />
    </div>
  );
}
