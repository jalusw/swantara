import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { TimesheetsSection } from "./_components/timesheets-section";

export default async function OrgTimesheetsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Attendance");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("timesheetsTitle")} description={t("timesheetsSubtitle")} />
      <TimesheetsSection orgId={id} />
    </div>
  );
}
