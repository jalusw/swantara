import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { ServiceSubNav } from "./_components/service-subnav";

export default async function ServiceLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("Service");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <Suspense fallback={null}>
        <ServiceSubNav />
      </Suspense>
      {children}
    </div>
  );
}
