import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ActivitiesSection } from "../_components/activities-section";

export default async function OrgCrmActivitiesPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Crm");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("activitiesTitle")} description={t("activitiesSubtitle")} />
      <ActivitiesSection orgId={id} />
    </div>
  );
}
