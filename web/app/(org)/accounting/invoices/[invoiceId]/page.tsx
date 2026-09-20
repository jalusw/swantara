import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InvoiceDetailSection } from "./_components/invoice-detail-section";

export default async function InvoiceDetailPage({
  params,
}: {
  params: Promise<{ invoiceId: string }>;
}) {
  const { invoiceId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/invoices"}>{"Back to invoices"}</BackLink>
      <InvoiceDetailSection orgId={id} invoiceId={invoiceId} />
    </div>
  );
}
