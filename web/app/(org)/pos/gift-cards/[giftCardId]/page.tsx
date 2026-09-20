import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { GiftCardDetail } from "./_components/gift-card-detail-section";

export default async function GiftCardDetailPage({
  params,
}: {
  params: Promise<{ giftCardId: string }>;
}) {
  const { giftCardId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/pos/gift-cards"}>{"Back to gift cards"}</BackLink>
      <GiftCardDetail orgId={id} giftCardId={giftCardId} />
    </div>
  );
}
