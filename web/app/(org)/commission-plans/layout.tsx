import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { CommissionsSubNav } from "./_components/commissions-subnav";

export default async function CommissionsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Commissions"} description={"Commission plans and earned entries."} />
      <Suspense fallback={null}>
        <CommissionsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
