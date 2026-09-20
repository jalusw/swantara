import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JobPositionsSection } from "./_components/job-positions-section";

export default async function OrgJobPositionsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Job positions"}
        description={"Manage job positions across departments."}
      />
      <JobPositionsSection orgId={id} />
    </div>
  );
}
