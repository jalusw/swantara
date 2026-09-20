import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { QualitySubNav } from "./_components/quality-subnav";

export default async function QualityLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Quality"} description={"Control points, checks, and alerts."} />
      <Suspense fallback={null}>
        <QualitySubNav />
      </Suspense>
      {children}
    </div>
  );
}
