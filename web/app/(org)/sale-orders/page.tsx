import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SaleOrdersSection } from "./_components/sale-orders-section";

export default async function OrgSaleOrdersPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Sale orders"}
        description={
          "Quotations and sale orders — from draft to done, with delivery and invoicing."
        }
      />
      <SaleOrdersSection orgId={id} />
    </div>
  );
}
