import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ProductsSection } from "./_components/products-section";

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: false,
  },
};

export default async function OrgProductsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Products",
  );
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("productsTitle")} description={t("productsDescription")} />
      <ProductsSection orgId={id} />
    </div>
  );
}
