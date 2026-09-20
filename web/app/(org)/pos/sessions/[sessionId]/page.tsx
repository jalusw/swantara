import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosSessionDetail } from "./_components/pos-session-detail-section";

export default async function OrgPosSessionDetailPage({
  params,
}: {
  params: Promise<{ sessionId: string }>;
}) {
  const { sessionId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/pos/sessions"}>{"All sessions"}</BackLink>
      <PosSessionDetail orgId={id} sessionId={sessionId} />
    </div>
  );
}
