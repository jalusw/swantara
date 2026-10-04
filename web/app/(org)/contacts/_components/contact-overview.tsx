"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Contacts");
  const items = [
    { id: "name", label: t("fieldName"), value: contact.name },
    {
      id: "type",
      label: t("fieldType"),
      value: contact.isOrganization ? t("typeOrganization") : t("typeIndividual"),
    },
    { id: "email", label: t("fieldEmail"), value: contact.email || "—" },
    { id: "phone", label: t("fieldPhone"), value: contact.phone || "—" },
    { id: "mobile", label: t("fieldMobile"), value: contact.mobile || "—" },
    { id: "website", label: t("fieldWebsite"), value: contact.website || "—" },
    { id: "taxId", label: t("fieldTaxId"), value: contact.taxId || "—" },
    { id: "industry", label: t("fieldIndustry"), value: contact.industry || "—" },
    {
      id: "currency",
      label: t("fieldCurrency"),
      value: contact.currencyCode || "—",
    },
    { id: "lang", label: t("fieldLanguage"), value: contact.lang || "—" },
  ];

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>{t("identity")}</CardTitle>
          <CardDescription>{t("identityDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DetailsList items={items} layout="grid" columns={3} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("rolesTitle")}</CardTitle>
          <CardDescription>{t("rolesDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">{t("rolesEmpty")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
