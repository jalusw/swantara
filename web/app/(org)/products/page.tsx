import type { Metadata } from "next";
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
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Products"}
        description={"Item catalog with variants, categories, and bills of materials."}
      />
      <ProductsSection orgId={id} />
    </div>
  );
}
