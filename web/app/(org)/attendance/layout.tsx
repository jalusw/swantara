import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { AttendanceSubNav } from "./_components/attendance-subnav";

export default async function AttendanceLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Presensi"} description={"Check-ins and timesheets for your team."} />
      <Suspense fallback={null}>
        <AttendanceSubNav />
      </Suspense>
      {children}
    </div>
  );
}
