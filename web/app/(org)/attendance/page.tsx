import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AttendanceSection } from "./_components/attendance-section";

export default async function OrgAttendancePage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Attendance");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <AttendanceSection orgId={id} />
    </div>
  );
}
