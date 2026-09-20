import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ReportsSubNav } from "./_components/reports-subnav";

export default async function ReportsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Reports"}
        description={"Financial and inventory reports for your organization."}
      />
      <Suspense fallback={null}>
        <ReportsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
