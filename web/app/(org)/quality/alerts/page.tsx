import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityAlertsSection } from "./_components/quality-alerts-section";

export default async function OrgQualityAlertsPage() {
  const t = await getTranslations("Quality");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("alertsTitle")} description={t("alertsDescription")} />
      <QualityAlertsSection orgId={id} />
    </div>
  );
}
