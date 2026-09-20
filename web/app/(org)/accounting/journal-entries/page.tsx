import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JournalEntriesSection } from "./_components/journal-entries-section";

export default async function OrgJournalEntriesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Journal Entries"}
        description={"View and manage journal entries with line-level detail."}
      />
      <JournalEntriesSection orgId={id} />
    </div>
  );
}
