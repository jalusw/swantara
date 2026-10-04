import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { SubscriptionDetail } from "./_components/subscription-detail-section";

export default async function SubscriptionDetailPage({
  params,
}: {
  params: Promise<{ subscriptionId: string }>;
}) {
  const t = await getTranslations("Subscriptions");
  const { subscriptionId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/subscriptions"}>{t("backToSubscriptions")}</BackLink>
      <SubscriptionDetail orgId={id} subscriptionId={subscriptionId} />
    </div>
  );
}
