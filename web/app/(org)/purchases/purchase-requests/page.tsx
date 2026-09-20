import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PurchaseRequestsSection } from "./_components/purchase-requests-section";

export default async function OrgPurchaseRequestsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Purchase requisitions"}
        description={
          "Internal purchase requests — from draft to approved, ready for QuoteRequest conversion."
        }
      />
      <PurchaseRequestsSection orgId={id} />
    </div>
  );
}
