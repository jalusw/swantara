import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { RmasSection } from "./_components/rmas-section";

export default async function OrgRmasPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Returns & RMA"}
        description={"Manage return merchandise authorizations."}
      />
      <RmasSection orgId={id} />
    </div>
  );
}
