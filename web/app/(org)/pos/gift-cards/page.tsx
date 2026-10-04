import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { GiftCardsSection } from "./_components/gift-cards-section";

export default async function GiftCardsPage() {
  const t = await getTranslations("Pos");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("giftCardsTitle")} description={t("giftCardsDescription")} />
      <GiftCardsSection orgId={id} />
    </div>
  );
}
