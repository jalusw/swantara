import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ServiceOrdersSection } from "./_components/service-orders-section";

export default async function ServiceOrdersPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Service orders"}
        description={"Manage service orders, repairs and maintenance work."}
      />
      <ServiceOrdersSection orgId={id} />
    </div>
  );
}
