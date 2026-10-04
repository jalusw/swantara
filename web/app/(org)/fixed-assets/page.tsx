import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { FixedAssetsSection } from "./_components/fixed-assets-section";

export default async function FixedAssetsPage() {
  const t = await getTranslations("FixedAssets");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <FixedAssetsSection orgId={id} />
    </div>
  );
}
