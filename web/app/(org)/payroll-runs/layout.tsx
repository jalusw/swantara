import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { PayrollSubNav } from "./_components/payroll-subnav";

export default async function PayrollLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Payroll"} description={"Salary rules, payroll runs, and payslips."} />
      <Suspense fallback={null}>
        <PayrollSubNav />
      </Suspense>
      {children}
    </div>
  );
}
