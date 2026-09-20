import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SubscriptionsSection } from "./_components/subscriptions-section";

export default async function SubscriptionsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Subscriptions"}
        description={"Manage subscriptions, recurring billing and revenue recognition."}
      />
      <SubscriptionsSection orgId={id} />
    </div>
  );
}
