import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { TimesheetsSection } from "./_components/timesheets-section";

export default async function OrgTimesheetsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Timesheets"} description={"Log hours worked on projects and tasks."} />
      <TimesheetsSection orgId={id} />
    </div>
  );
}
