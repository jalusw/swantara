import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { StatementDetailSection } from "./_components/statement-detail-section";

export default async function StatementDetailPage({
  params,
}: {
  params: Promise<{ statementId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const { statementId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/bank-statements"}>{t("backToStatements")}</BackLink>
      <StatementDetailSection orgId={id} statementId={statementId} />
    </div>
  );
}
