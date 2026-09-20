import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { WarehousesSection } from "./_components/warehouses-section";

export default async function OrgWarehousesPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Warehouses"}
        description={"Manage warehouse locations and their usage types."}
      />
      <WarehousesSection orgId={id} />
    </div>
  );
}
