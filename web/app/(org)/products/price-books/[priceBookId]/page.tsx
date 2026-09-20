import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PriceBookDetail } from "./_components/price-book-detail-section";

export default async function OrgPriceBookDetailPage({
  params,
}: {
  params: Promise<{ priceBookId: string }>;
}) {
  const { priceBookId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/products/price_books"}>{"Back to price_books"}</BackLink>
      <PriceBookDetail orgId={id} priceBookId={priceBookId} />
    </div>
  );
}
