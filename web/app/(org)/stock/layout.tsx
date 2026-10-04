import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { StockSubNav } from "./_components/stock-subnav";

export default async function StockLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Stok"}
        description={"On-hand quantities, warehouses, and stock movements across locations."}
      />
      <Suspense fallback={null}>
        <StockSubNav />
      </Suspense>
      {children}
    </div>
  );
}
