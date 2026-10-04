import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AssetCategoriesSection } from "./_components/asset-categories-section";

export default async function AssetCategoriesPage() {
  const t = await getTranslations("FixedAssets");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("categoriesTitle")} description={t("categoriesDescription")} />
      <AssetCategoriesSection orgId={id} />
    </div>
  );
}
