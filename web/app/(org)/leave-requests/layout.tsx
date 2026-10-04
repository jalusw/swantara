import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { LeaveSubNav } from "./_components/leave-subnav";

export default async function LeaveLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Cuti"}
        description={"Leave types and employee leave requests with approvals."}
      />
      <Suspense fallback={null}>
        <LeaveSubNav />
      </Suspense>
      {children}
    </div>
  );
}
