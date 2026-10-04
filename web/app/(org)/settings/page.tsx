import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { FinancialCard } from "./_components/financial-card";
import { ModulesCard } from "./_components/modules-card";
import { ProfileCard } from "./_components/profile-card";

export default async function OrgGeneralSettingsPage() {
  const t = await getTranslations("Settings");
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />

      <ProfileCard orgId={id} />

      <FinancialCard orgId={id} />

      <ModulesCard />
    </div>
  );
}
