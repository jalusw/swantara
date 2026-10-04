import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InboundCostsSection } from "./_components/inbound-costs-section";

export default async function InboundCostsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Purchases");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("inboundCostsTitle")} description={t("inboundCostsSubtitle")} />
      <InboundCostsSection orgId={id} />
    </div>
  );
}
