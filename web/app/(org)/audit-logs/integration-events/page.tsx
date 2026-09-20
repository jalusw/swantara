import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { IntegrationEventsSection } from "./_components/integration-events-section";

export default async function OrgIntegrationEventsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Integration events"}
        description={"Monitor integration events and delivery status."}
      />
      <IntegrationEventsSection orgId={id} />
    </div>
  );
}
