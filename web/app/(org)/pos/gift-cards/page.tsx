import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { GiftCardsSection } from "./_components/gift-cards-section";

export default async function GiftCardsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Gift cards"} description={"Issue, redeem and manage gift cards."} />
      <GiftCardsSection orgId={id} />
    </div>
  );
}
