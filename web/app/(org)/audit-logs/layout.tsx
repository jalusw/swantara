import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ActivitySubNav } from "./_components/activity-subnav";

export default async function ActivityLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Activity"} description={"Audit trail and integration events."} />
      <Suspense fallback={null}>
        <ActivitySubNav />
      </Suspense>
      {children}
    </div>
  );
}
