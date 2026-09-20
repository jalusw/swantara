import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { TeamSubNav } from "./_components/team-subnav";

export default async function TeamLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Team"} description={"Employees, departments, and job positions."} />
      <Suspense fallback={null}>
        <TeamSubNav />
      </Suspense>
      {children}
    </div>
  );
}
