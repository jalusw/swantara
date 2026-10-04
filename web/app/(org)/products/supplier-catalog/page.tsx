import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SupplierCatalogSection } from "./_components/supplier-catalog-section";

export default async function OrgSupplierCatalogPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Products",
  );
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("supplierCatalogTitle")} description={t("supplierCatalogDescription")} />
      <SupplierCatalogSection orgId={id} />
    </div>
  );
}
