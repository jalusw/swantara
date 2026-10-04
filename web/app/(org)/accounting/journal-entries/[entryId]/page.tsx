import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JournalEntryDetailSection } from "./_components/journal-entry-detail-section";

export default async function MoveDetailPage({ params }: { params: Promise<{ entryId: string }> }) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const { entryId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/journal-entries"}>{t("backToJournalEntries")}</BackLink>
      <JournalEntryDetailSection orgId={id} entryId={entryId} />
    </div>
  );
}
