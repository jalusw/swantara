import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { EquipmentsSection } from "./_components/equipments-section";

export default async function EquipmentsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Equipments"} description={"Manage equipment and asset tracking."} />
      <EquipmentsSection orgId={id} />
    </div>
  );
}
