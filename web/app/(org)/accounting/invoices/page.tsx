import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InvoicesSection } from "./_components/invoices-section";

export default async function OrgInvoicesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Invoices"} description={"Manage customer invoices and supplier bills."} />
      <InvoicesSection orgId={id} />
    </div>
  );
}
