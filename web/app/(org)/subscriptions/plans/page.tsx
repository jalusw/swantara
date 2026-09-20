import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SubscriptionPlansSection } from "./_components/subscription-plans-section";

export default async function SubscriptionPlansPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Subscription plans"}
        description={"Configure recurring billing intervals and counts."}
      />
      <SubscriptionPlansSection orgId={id} />
    </div>
  );
}
