import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PipelineSection } from "./_components/pipeline-section";

const LEGACY_TAB_REDIRECT: Record<string, string> = {
  leads: "leads",
  opportunities: "opportunities",
  activities: "activities",
};

export default async function OrgCrmPage({
  searchParams,
}: {
  searchParams?: Promise<{ tab?: string | string[] }>;
}) {
  const id = String(await requireActiveOrgId());
  const sp = searchParams ? await searchParams : {};
  const rawTabValue = sp?.tab;
  const rawTab = Array.isArray(rawTabValue) ? rawTabValue[0] : rawTabValue;
  if (rawTab && LEGACY_TAB_REDIRECT[rawTab]) {
    redirect(`/crm/${LEGACY_TAB_REDIRECT[rawTab]}`);
  }
  const t = await getTranslations("Crm");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("pipelineTitle")} description={t("pipelineSubtitle")} />
      <PipelineSection orgId={id} />
    </div>
  );
}
