import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ServiceContractsSection } from "./_components/service-contracts-section";

export default async function ServiceContractsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Service contracts"}
        description={"Manage service contracts and SLA agreements."}
      />
      <ServiceContractsSection orgId={id} />
    </div>
  );
}
