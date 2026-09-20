import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { PosSubNav } from "./_components/pos-subnav";

export default async function PosLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Point of sale"}
        description={"Registers, sessions, orders, and gift cards."}
      />
      <Suspense fallback={null}>
        <PosSubNav />
      </Suspense>
      {children}
    </div>
  );
}
