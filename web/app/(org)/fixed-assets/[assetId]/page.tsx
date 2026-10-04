import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { FixedAssetDetail } from "./_components/fixed-asset-detail-section";

export default async function FixedAssetDetailPage({
  params,
}: {
  params: Promise<{ assetId: string }>;
}) {
  const t = await getTranslations("FixedAssets");
  const { assetId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/fixed-assets"}>{t("backToAssets")}</BackLink>
      <FixedAssetDetail orgId={id} assetId={assetId} />
    </div>
  );
}
