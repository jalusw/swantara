import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SupplierQuoteRequestsSection } from "./_components/supplier-quote-requests-section";

export default async function OrgSupplierQuoteRequestsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Purchase Quote Requests"}
        description={
          "Request for Quotations — send to suppliers, collect quotes, and convert to purchase orders."
        }
      />
      <SupplierQuoteRequestsSection orgId={id} />
    </div>
  );
}
