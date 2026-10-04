import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ProductsSubNav } from "./_components/products-subnav";

export default async function ProductsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Produk"}
        description={"Katalog barang dengan varian, kategori, dan resep."}
      />
      <Suspense fallback={null}>
        <ProductsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
