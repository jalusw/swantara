import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { OpportunitiesSection } from "../_components/opportunities-section";

export default async function OrgCrmOpportunitiesPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Crm");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("opportunitiesTitle")} description={t("opportunitiesSubtitle")} />
      <OpportunitiesSection orgId={id} />
    </div>
  );
}
