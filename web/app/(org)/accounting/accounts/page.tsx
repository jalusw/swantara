import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AccountsSection } from "./_components/accounts-section";

export default async function OrgAccountsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Chart of Accounts"}
        description={"Manage your chart of accounts with account codes, types, and hierarchy."}
      />
      <AccountsSection orgId={id} />
    </div>
  );
}
