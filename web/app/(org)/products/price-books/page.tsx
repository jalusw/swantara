import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PriceBooksSection } from "./_components/price-books-section";

export default async function OrgPriceBooksPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"PriceBooks"}
        description={"Manage pricing rules for products across customers and channels."}
      />
      <PriceBooksSection orgId={id} />
    </div>
  );
}
