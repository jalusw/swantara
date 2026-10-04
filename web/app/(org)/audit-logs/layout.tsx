import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ActivitySubNav } from "./_components/activity-subnav";

export default async function ActivityLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("AuditLogs");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("layoutTitle")} description={t("layoutDescription")} />
      <Suspense fallback={null}>
        <ActivitySubNav />
      </Suspense>
      {children}
    </div>
  );
}
