import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MrpSection } from "./_components/planning-section";

export default async function OrgMrpPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Planning Planning"} description={"Material Requirements Planning"} />
      <MrpSection orgId={id} />
    </div>
  );
}
