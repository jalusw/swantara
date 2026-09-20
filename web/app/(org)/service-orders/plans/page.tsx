import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MaintenancePlansSection } from "./_components/maintenance-plans-section";

export default async function MaintenancePlansPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Maintenance plans"}
        description={"Schedule preventive maintenance for equipment."}
      />
      <MaintenancePlansSection orgId={id} />
    </div>
  );
}
