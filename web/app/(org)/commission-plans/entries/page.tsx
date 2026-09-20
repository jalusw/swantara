import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CommissionEntriesSection } from "./_components/commission-entries-section";

export default async function CommissionEntriesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Commission entries"}
        description={"View and manage commission accruals and payments."}
      />
      <CommissionEntriesSection orgId={id} />
    </div>
  );
}
