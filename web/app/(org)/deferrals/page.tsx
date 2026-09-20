import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DeferralsSection } from "./_components/deferrals-section";

export default async function DeferralsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Deferrals"}
        description={"Manage deferral schedules and revenue/expense recognition."}
      />
      <DeferralsSection orgId={id} />
    </div>
  );
}
