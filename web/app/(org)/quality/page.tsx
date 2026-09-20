import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityChecksSection } from "./checks/_components/quality-checks-section";

export default async function OrgQualityPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Quality checks"}
        description={"Record and review quality check results."}
      />
      <QualityChecksSection orgId={id} />
    </div>
  );
}
