import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ContactDetail } from "../_components/contact-detail-section";

export default async function OrgContactDetailPage({
  params,
}: {
  params: Promise<{ contactId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Contacts",
  );
  const { contactId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/contacts"}>{t("backToContacts")}</BackLink>
      <ContactDetail orgId={id} contactId={contactId} />
    </div>
  );
}
