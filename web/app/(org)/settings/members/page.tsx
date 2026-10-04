import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";

import { MembersSection } from "./_components/members-section";

export default async function OrgMembersPage() {
  const t = await getTranslations("Settings");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("membersTitle")} description={t("membersDescription")} />
      <MembersSection />
    </div>
  );
}
