import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosRegister } from "./_components/pos-register-section";

export default async function PosRegisterPage({
  params,
}: {
  params: Promise<{ sessionId: string }>;
}) {
  const { sessionId } = await params;
  const id = String(await requireActiveOrgId());
  return <PosRegister orgId={id} sessionId={sessionId} />;
}
