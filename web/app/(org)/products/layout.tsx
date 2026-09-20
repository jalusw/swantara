import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ProductsSubNav } from "./_components/products-subnav";

export default async function ProductsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Products"}
        description={"Item catalog with variants, categories, and bills of materials."}
      />
      <Suspense fallback={null}>
        <ProductsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
