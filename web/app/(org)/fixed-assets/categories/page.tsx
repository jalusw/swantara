import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AssetCategoriesSection } from "./_components/asset-categories-section";

export default async function AssetCategoriesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Asset categories"}
        description={"Configure depreciation methods and GL accounts for asset classes."}
      />
      <AssetCategoriesSection orgId={id} />
    </div>
  );
}
