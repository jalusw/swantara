import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { TaxesSection } from "./_components/taxes-section";

export default async function OrgTaxesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Taxes"} description={"Configure tax rates and tax groups."} />
      <TaxesSection orgId={id} />
    </div>
  );
}
