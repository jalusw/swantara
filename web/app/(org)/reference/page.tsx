import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ReferenceDataTabs } from "./_components/reference-data-tabs";

export default async function OrgReferencePage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Reference data"}
        description={
          "Shared data every module depends on: currencies, exchange rates, units of measure, dimension accounts, and payment terms."
        }
      />
      <ReferenceDataTabs orgId={id} />
    </div>
  );
}
