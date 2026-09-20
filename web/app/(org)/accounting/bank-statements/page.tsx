import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { BankStatementsSection } from "./_components/bank-statements-section";

export default async function OrgBankStatementsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Bank Statements"}
        description={"Manage bank statements and reconciliation."}
      />
      <BankStatementsSection orgId={id} />
    </div>
  );
}
