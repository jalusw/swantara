import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ServiceContractsSection } from "./_components/service-contracts-section";

export default async function ServiceContractsPage() {
  const t = await getTranslations("Service");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("contractsTitle")} description={t("contractsDescription")} />
      <ServiceContractsSection orgId={id} />
    </div>
  );
}
