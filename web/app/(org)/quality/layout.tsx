import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { QualitySubNav } from "./_components/quality-subnav";

export default async function QualityLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("Quality");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <Suspense fallback={null}>
        <QualitySubNav />
      </Suspense>
      {children}
    </div>
  );
}
