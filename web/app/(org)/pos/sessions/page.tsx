import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosSessionsSection } from "./_components/pos-sessions-section";

export default async function OrgPosSessionsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"POS Sessions"} description={"Open, manage and close POS sessions."} />
      <PosSessionsSection orgId={id} />
    </div>
  );
}
