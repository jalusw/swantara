import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CategoryTreeSection } from "./_components/category-tree-section";

export default async function OrgProductCategoriesPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Products",
  );
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("categoriesTitle")} description={t("categoriesDescription")} />
      <CategoryTreeSection orgId={id} />
    </div>
  );
}
