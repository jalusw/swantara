import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ProductDetail } from "../_components/product-detail-section";

export default async function OrgProductDetailPage({
  params,
}: {
  params: Promise<{ itemId: string }>;
}) {
  const { itemId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/products"}>{"Back to products"}</BackLink>
      <ProductDetail orgId={id} itemId={itemId} />
    </div>
  );
}
