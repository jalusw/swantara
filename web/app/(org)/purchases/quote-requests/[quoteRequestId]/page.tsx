import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SupplierQuoteRequestDetail } from "./_components/supplier-quote-request-detail-section";

export default async function OrgSupplierQuoteRequestDetailPage({
  params,
}: {
  params: Promise<{ quoteRequestId: string }>;
}) {
  const { quoteRequestId } = await params;
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/purchases/quoteRequests"}>{t("backToQuoteRequests")}</BackLink>
      <SupplierQuoteRequestDetail orgId={id} quoteRequestId={quoteRequestId} />
    </div>
  );
}
