import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { AssetsSubNav } from "./_components/assets-subnav";

export default async function AssetsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Assets"}
        description={"Fixed assets and their categories with depreciation."}
      />
      <Suspense fallback={null}>
        <AssetsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
