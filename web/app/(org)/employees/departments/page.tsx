import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DepartmentsSection } from "./_components/departments-section";

export default async function OrgDepartmentsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Departments"}
        description={"Manage organizational departments and structure."}
      />
      <DepartmentsSection orgId={id} />
    </div>
  );
}
