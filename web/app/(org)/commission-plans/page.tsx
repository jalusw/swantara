import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CommissionPlansSection } from "./_components/commission-plans-section";

export default async function CommissionPlansPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Commission plans"}
        description={"Manage commission plans, rules and assignments."}
      />
      <CommissionPlansSection orgId={id} />
    </div>
  );
}
