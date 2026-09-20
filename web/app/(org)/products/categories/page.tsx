import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CategoryTreeSection } from "./_components/category-tree-section";

export default async function OrgProductCategoriesPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Item categories"}
        description={"Organize products into a tree with accounting defaults per category."}
      />
      <CategoryTreeSection orgId={id} />
    </div>
  );
}
