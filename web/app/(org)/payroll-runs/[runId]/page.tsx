import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PayrollRunDetail } from "./_components/payroll-run-detail-section";

export default async function PayrollRunDetailPage({
  params,
}: {
  params: Promise<{ runId: string }>;
}) {
  const { runId } = await params;
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Payroll");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/payroll-runs"}>{t("backToRuns")}</BackLink>
      <PayrollRunDetail orgId={id} runId={runId} />
    </div>
  );
}
