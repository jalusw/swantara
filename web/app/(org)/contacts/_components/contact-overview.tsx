"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DetailsList } from "@/components/details-list";
import type { Contact } from "@/lib/services/swantara";

export function ContactOverview({
  orgId: _orgId,
  contactId: _contactId,
  contact,
  onRefetch: _onRefetch,
}: {
  orgId: string;
  contactId: string;
  contact: Contact;
  onRefetch: () => void;
}) {
  const items = [
    { id: "name", label: "Name", value: contact.name },
    {
      id: "type",
      label: "Type",
      value: contact.isOrganization ? "Organization" : "Individual",
    },
    { id: "email", label: "Email", value: contact.email || "—" },
    { id: "phone", label: "Phone", value: contact.phone || "—" },
    { id: "mobile", label: "Mobile", value: contact.mobile || "—" },
    { id: "website", label: "Website", value: contact.website || "—" },
    { id: "taxId", label: "Tax ID", value: contact.taxId || "—" },
    { id: "industry", label: "Industry", value: contact.industry || "—" },
    {
      id: "currency",
      label: "Currency",
      value: contact.currencyCode || "—",
    },
    { id: "lang", label: "Language", value: contact.lang || "—" },
  ];

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>{"Identity"}</CardTitle>
          <CardDescription>{"Core details about this contact."}</CardDescription>
        </CardHeader>
        <CardContent>
          <DetailsList items={items} layout="grid" columns={3} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{"Roles"}</CardTitle>
          <CardDescription>{"Enable the role extensions this contact uses."}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            {"No roles enabled. This contact is used as a contact only."}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
