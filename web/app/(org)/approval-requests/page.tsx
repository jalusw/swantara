import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";

import { ApprovalRequestsSection } from "./_components/approval-requests-section";

export default async function OrgApprovalRequestsPage() {
  const t = await getTranslations("ApprovalRequests");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <ApprovalRequestsSection />
    </div>
  );
}
