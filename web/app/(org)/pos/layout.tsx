import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { PosSubNav } from "./_components/pos-subnav";

export default async function PosLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("Pos");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <Suspense fallback={null}>
        <PosSubNav />
      </Suspense>
      {children}
    </div>
  );
}
