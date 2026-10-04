import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { RmaDetailSection } from "./_components/rma-detail-section";

export default async function RmaDetailPage({ params }: { params: Promise<{ rmaId: string }> }) {
  const { rmaId } = await params;
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Sales");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/sale-orders/returns"}>{t("backToRmas")}</BackLink>
      <RmaDetailSection orgId={id} rmaId={rmaId} />
    </div>
  );
}
