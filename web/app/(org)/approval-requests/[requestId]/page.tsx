import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ApprovalRequestDetail } from "./_components/approval-request-detail-section";

export default async function ApprovalRequestDetailPage({
  params,
}: {
  params: Promise<{ requestId: string }>;
}) {
  const t = await getTranslations("ApprovalRequests");
  const { requestId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/approval-requests"}>{t("backToRequests")}</BackLink>
      <ApprovalRequestDetail orgId={id} requestId={requestId} />
    </div>
  );
}
