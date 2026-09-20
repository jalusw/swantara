import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { CrmSubNav } from "./_components/crm-subnav";

export default async function CrmLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"CRM"}
        description={
          "Leads, opportunities and the sales pipeline — from first contact to won deal."
        }
      />
      <Suspense fallback={null}>
        <CrmSubNav />
      </Suspense>
      {children}
    </div>
  );
}
