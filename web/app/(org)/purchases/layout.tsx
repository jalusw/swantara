import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { PurchasesSubNav } from "./_components/purchases-subnav";

export default async function PurchasesLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Purchases"}
        description={"Requisitions, quotes, orders, and landed costs — procure to pay."}
      />
      <Suspense fallback={null}>
        <PurchasesSubNav />
      </Suspense>
      {children}
    </div>
  );
}
