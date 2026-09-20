import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SupplierCatalogSection } from "./_components/supplier-catalog-section";

export default async function OrgSupplierCatalogPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Supplier catalog"}
        description={"Item-to-supplier pricing, lead times, and priority."}
      />
      <SupplierCatalogSection orgId={id} />
    </div>
  );
}
