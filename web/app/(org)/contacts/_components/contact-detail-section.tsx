"use client";

import { Pencil } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { RecordLayout } from "@/components/record-layout";
import { useContactQuery } from "@/lib/hooks/use-contact-query";

import { ContactAddresses } from "./contact-addresses";
import { ContactBankAccounts } from "./contact-bank-accounts";
import { ContactDefaults } from "./contact-defaults";
import { ContactFormDialog } from "./contact-form-dialog";
import { ContactOverview } from "./contact-overview";

export function ContactDetail({ orgId, contactId }: { orgId: string; contactId: string }) {
  const [editOpen, setEditOpen] = useState(false);

  const contactQuery = useContactQuery(orgId, contactId);
  const contact = contactQuery.data;

  if (contactQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!contact) {
    return <p className="text-sm text-muted-foreground">{"Contact not found."}</p>;
  }

  const displayName = contact.displayName || contact.name;

  function handleEditSave() {
    setEditOpen(false);
    toast.success("Contact saved.");
    void contactQuery.refetch();
  }

  const status = (
    <Badge variant={contact.active ? "default" : "outline"}>
      {contact.active ? "Active" : "Inactive"}
    </Badge>
  );

  return (
    <>
      <RecordLayout
        breadcrumbItems={[{ label: "Contacts", href: "/contacts" }, { label: displayName }]}
        title={displayName}
        description={contact.isOrganization ? contact.name : undefined}
        status={status}
        actions={
          <Button size="sm" variant="outline" onClick={() => setEditOpen(true)}>
            <Pencil />
            <span>{"Edit"}</span>
          </Button>
        }
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <ContactOverview
                orgId={orgId}
                contactId={contactId}
                contact={contact}
                onRefetch={() => void contactQuery.refetch()}
              />
            ),
          },
          {
            id: "addresses",
            label: "Addresses",
            content: (
              <ContactAddresses
                orgId={orgId}
                contactId={contactId}
                onRefetch={() => void contactQuery.refetch()}
              />
            ),
          },
          {
            id: "bank-accounts",
            label: "Bank accounts",
            content: (
              <ContactBankAccounts
                orgId={orgId}
                contactId={contactId}
                onRefetch={() => void contactQuery.refetch()}
              />
            ),
          },
          {
            id: "defaults",
            label: "Defaults",
            content: (
              <ContactDefaults
                orgId={orgId}
                contactId={contactId}
                contact={contact}
                onRefetch={() => void contactQuery.refetch()}
              />
            ),
          },
        ]}
      />
      <ContactFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        orgId={orgId}
        initial={contact}
        onSave={handleEditSave}
      />
    </>
  );
}
