import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JobPositionsSection } from "./_components/job-positions-section";

export default async function OrgJobPositionsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Employees");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("jobPositionsTitle")} description={t("jobPositionsSubtitle")} />
      <JobPositionsSection orgId={id} />
    </div>
  );
}
