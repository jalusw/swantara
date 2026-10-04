import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { AssetsSubNav } from "./_components/assets-subnav";

export default async function AssetsLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("FixedAssets");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("layoutTitle")} description={t("layoutDescription")} />
      <Suspense fallback={null}>
        <AssetsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
