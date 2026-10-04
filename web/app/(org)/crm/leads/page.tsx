import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeadsSection } from "../_components/leads-section";

export default async function OrgCrmLeadsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Crm");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("leadsTitle")} description={t("leadsSubtitle")} />
      <LeadsSection orgId={id} />
    </div>
  );
}
