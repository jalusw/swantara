import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { PageHeader } from "@/components/page-header";
import { SubscriptionsSubNav } from "./_components/subscriptions-subnav";

export default async function SubscriptionsLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations("Subscriptions");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("layoutDescription")} />
      <Suspense fallback={null}>
        <SubscriptionsSubNav />
      </Suspense>
      {children}
    </div>
  );
}
