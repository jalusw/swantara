import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InvoiceDetailSection } from "./_components/invoice-detail-section";

export default async function InvoiceDetailPage({
  params,
}: {
  params: Promise<{ invoiceId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const { invoiceId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/invoices"}>{t("backToInvoices")}</BackLink>
      <InvoiceDetailSection orgId={id} invoiceId={invoiceId} />
    </div>
  );
}
