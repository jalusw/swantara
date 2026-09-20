import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SalaryRulesSection } from "./_components/salary-rules-section";

export default async function OrgSalaryRulesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Salary rules"}
        description={"Configure earnings and deduction rules for payroll computation."}
      />
      <SalaryRulesSection orgId={id} />
    </div>
  );
}
