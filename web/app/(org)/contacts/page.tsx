import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ContactsSection } from "./_components/contacts-section";

export default async function OrgContactsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Contacts"}
        description={"Unified identity for customers, suppliers, and employees."}
      />
      <ContactsSection orgId={id} />
    </div>
  );
}
