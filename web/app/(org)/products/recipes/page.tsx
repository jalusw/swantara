import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { BomsSection } from "./_components/recipes-section";

export default async function OrgBomsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Bills of materials"}
        description={
          "Recipes that define how manufactured, kit, and subcontracted products are assembled."
        }
      />
      <BomsSection orgId={id} />
    </div>
  );
}
