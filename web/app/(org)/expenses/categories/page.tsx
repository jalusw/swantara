import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ExpenseCategoriesSection } from "./_components/expense-categories-section";

export default async function ExpenseCategoriesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Expense categories"}
        description={"Manage expense categories and default accounts."}
      />
      <ExpenseCategoriesSection orgId={id} />
    </div>
  );
}
