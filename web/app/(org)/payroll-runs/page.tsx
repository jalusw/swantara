import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PayrollRunsSection } from "./_components/payroll-runs-section";

export default async function OrgPayrollRunsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Payroll runs"}
        description={"Create and process payroll runs for your organization."}
      />
      <PayrollRunsSection orgId={id} />
    </div>
  );
}
