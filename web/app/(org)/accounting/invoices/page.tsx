import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InvoicesSection } from "./_components/invoices-section";

export default async function OrgInvoicesPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("invoicesTitle")} description={t("invoicesDescription")} />
      <InvoicesSection orgId={id} />
    </div>
  );
}
