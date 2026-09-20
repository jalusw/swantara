import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { TaxYearsSection } from "./_components/tax-years-section";

export default async function OrgTaxYearsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Fiscal Years"}
        description={"Manage tax years and their open/closed status."}
      />
      <TaxYearsSection orgId={id} />
    </div>
  );
}
