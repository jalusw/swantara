import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ServiceSubNav } from "./_components/service-subnav";

export default async function ServiceLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Service"}
        description={"Equipment, contracts, maintenance plans, and service orders."}
      />
      <Suspense fallback={null}>
        <ServiceSubNav />
      </Suspense>
      {children}
    </div>
  );
}
