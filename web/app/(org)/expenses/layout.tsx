import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ExpensesSubNav } from "./_components/expenses-subnav";

export default async function ExpensesLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Biaya"}
        description={"Expense categories and employee expense reports."}
      />
      <Suspense fallback={null}>
        <ExpensesSubNav />
      </Suspense>
      {children}
    </div>
  );
}
