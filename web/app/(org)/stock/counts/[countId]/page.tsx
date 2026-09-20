import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { CountDetail } from "./_components/count-detail-section";

export default async function CountDetailPage({
  params,
}: {
  params: Promise<{ countId: string }>;
}) {
  const { countId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/stock/counts"}>{"Back to counts"}</BackLink>
      <CountDetail orgId={id} countId={countId} />
    </div>
  );
}
