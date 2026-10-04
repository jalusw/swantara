import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SubscriptionsSection } from "./_components/subscriptions-section";

export default async function SubscriptionsPage() {
  const t = await getTranslations("Subscriptions");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <SubscriptionsSection orgId={id} />
    </div>
  );
}
