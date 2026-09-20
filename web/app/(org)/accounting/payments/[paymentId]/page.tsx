import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PaymentDetailSection } from "./_components/payment-detail-section";

export default async function PaymentDetailPage({
  params,
}: {
  params: Promise<{ paymentId: string }>;
}) {
  const { paymentId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/accounting/payments"}>{"Back to payments"}</BackLink>
      <PaymentDetailSection orgId={id} paymentId={paymentId} />
    </div>
  );
}
