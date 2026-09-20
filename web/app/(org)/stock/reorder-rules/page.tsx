import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ReorderRulesSection } from "./_components/reorder-rules-section";

export default async function OrgReorderRulesPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Reorder rules"}
        description={"Define minimum and maximum stock levels for automatic replenishment."}
      />
      <ReorderRulesSection orgId={id} />
    </div>
  );
}
