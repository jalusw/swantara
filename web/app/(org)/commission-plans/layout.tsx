import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { CommissionsSubNav } from "./_components/commissions-subnav";

export default async function CommissionsLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("Commissions");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <Suspense fallback={null}>
        <CommissionsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
