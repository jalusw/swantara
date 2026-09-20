import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ExpensesSection } from "./_components/expenses-section";

export default async function ExpensesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Expenses"}
        description={"Manage employee expense reports and reimbursements."}
      />
      <ExpensesSection orgId={id} />
    </div>
  );
}
