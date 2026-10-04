import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SupplierQuoteRequestsSection } from "./_components/supplier-quote-requests-section";

export default async function OrgSupplierQuoteRequestsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("quoteRequestsTitle")} description={t("quoteRequestsSubtitle")} />
      <SupplierQuoteRequestsSection orgId={id} />
    </div>
  );
}
