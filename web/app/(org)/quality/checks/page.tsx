import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { QualityChecksSection } from "./_components/quality-checks-section";

export default async function OrgQualityChecksPage() {
  const t = await getTranslations("Quality");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("checksTitle")} description={t("checksDescription")} />
      <QualityChecksSection orgId={id} />
    </div>
  );
}
