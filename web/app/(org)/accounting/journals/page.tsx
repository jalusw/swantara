import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JournalsSection } from "./_components/journals-section";

export default async function OrgJournalsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Journals"}
        description={"Manage journals for recording financial transactions."}
      />
      <JournalsSection orgId={id} />
    </div>
  );
}
