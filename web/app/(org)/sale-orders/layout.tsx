import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { SalesSubNav } from "./_components/sales-subnav";

export default async function SalesLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Penjualan"} description={"Customers, sale orders, and returns."} />
      <Suspense fallback={null}>
        <SalesSubNav />
      </Suspense>
      {children}
    </div>
  );
}
