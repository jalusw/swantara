import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityPointsSection } from "./_components/quality-points-section";

export default async function OrgQualityPointsPage() {
  const t = await getTranslations("Quality");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("pointsTitle")} description={t("pointsDescription")} />
      <QualityPointsSection orgId={id} />
    </div>
  );
}
