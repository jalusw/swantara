import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AttendanceSection } from "./_components/attendance-section";

export default async function OrgAttendancePage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Attendance"}
        description={"Track employee check-in and check-out times."}
      />
      <AttendanceSection orgId={id} />
    </div>
  );
}
