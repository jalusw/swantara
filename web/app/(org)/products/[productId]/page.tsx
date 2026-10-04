import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ProductDetail } from "../_components/product-detail-section";

export default async function OrgProductDetailPage({
  params,
}: {
  params: Promise<{ itemId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Products",
  );
  const { itemId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/products"}>{t("backToProducts")}</BackLink>
      <ProductDetail orgId={id} itemId={itemId} />
    </div>
  );
}
