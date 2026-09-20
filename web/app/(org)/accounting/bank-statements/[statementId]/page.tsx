import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { StatementDetailSection } from "./_components/statement-detail-section";

export default async function StatementDetailPage({
  params,
}: {
  params: Promise<{ statementId: string }>;
}) {
  const { statementId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/bank-statements"}>{"Back to statements"}</BackLink>
      <StatementDetailSection orgId={id} statementId={statementId} />
    </div>
  );
}
