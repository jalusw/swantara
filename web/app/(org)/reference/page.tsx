import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ReferenceDataTabs } from "./_components/reference-data-tabs";

export default async function OrgReferencePage() {
  const t = await getTranslations("Reference");
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <ReferenceDataTabs orgId={id} />
    </div>
  );
}
