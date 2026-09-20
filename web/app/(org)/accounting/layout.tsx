import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { AccountingSubNav } from "./_components/accounting-subnav";

export default async function AccountingLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Accounting"}
        description={"Chart of accounts, journals, taxes, and financial transactions."}
      />
      <Suspense fallback={null}>
        <AccountingSubNav />
      </Suspense>
      {children}
    </div>
  );
}
