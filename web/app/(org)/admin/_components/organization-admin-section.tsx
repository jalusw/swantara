"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useMeOrganizationsQuery } from "@/lib/hooks/use-me-query";
import { useActiveOrg, useOrganizationId } from "@/lib/hooks/use-org-context";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";

type ParentOption = { id: number; name: string };

const baseCurrencies = [
  { code: "USD", name: "US Dollar (USD)" },
  { code: "IDR", name: "Indonesian Rupiah (IDR)" },
  { code: "EUR", name: "Euro (EUR)" },
  { code: "SGD", name: "Singapore Dollar (SGD)" },
];

const countries = [
  { code: "US", name: "United States" },
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapore" },
  { code: "DE", name: "Germany" },
  { code: "NL", name: "Netherlands" },
  { code: "GB", name: "United Kingdom" },
];

const timezones = ["UTC", "Asia/Jakarta", "Asia/Singapore", "America/New_York", "Europe/Amsterdam"];

function useOrganizationFormSchema() {
  return z.object({
    name: z.string().min(1, "Name"),
    legalName: z.string(),
    parentId: z.string(),
    baseCurrency: z.string().min(1, "Base currency"),
    countryCode: z.string(),
    taxId: z.string(),
    timezone: z.string().min(1, "Timezone"),
    taxYearStartMonth: z.string().min(1, "Tax year starts in"),
  });
}
type OrganizationFormValues = z.infer<ReturnType<typeof useOrganizationFormSchema>>;

export function OrganizationAdminSection({ orgId }: { orgId: string }) {
  const activeOrg = useActiveOrg();
  useOrganizationId();
  const orgListQuery = useMeOrganizationsQuery();
  const { data: orgList } = orgListQuery;

  const orgData = (orgList?.organizations ?? []).find(
    (org) => org.id === activeOrg?.id || String(org.id) === orgId,
  );

  const organization = orgData
    ? {
        id: orgData.id,
        name: orgData.name,
        legalName: orgData.name,
        parentId: null as number | null,
        baseCurrency: "USD",
        countryCode: "US",
        taxId: "",
        timezone: "UTC",
        taxYearStartMonth: 1,
      }
    : null;

  const parents: ParentOption[] = (orgList?.organizations ?? [])
    .filter((org) => String(org.id) !== orgId)
    .map((org) => ({ id: org.id, name: org.name }));

  const schema = useOrganizationFormSchema();

  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(schema),
    defaultValues: organization
      ? {
          name: organization.name,
          legalName: organization.legalName,
          parentId: organization.parentId == null ? "" : String(organization.parentId),
          baseCurrency: organization.baseCurrency,
          countryCode: organization.countryCode,
          taxId: organization.taxId,
          timezone: organization.timezone,
          taxYearStartMonth: String(organization.taxYearStartMonth),
        }
      : undefined,
  });

  const isActiveOrg = (Number(orgId) || 0) === activeOrg?.id;
  const disableDisabled = isActiveOrg;

  function handleSave(values: OrganizationFormValues) {
    if (!organization) return;
    void getSwantaraService()
      .organizations.update(Number(orgId) || 0, {
        name: values.name,
        legalName: values.legalName || "",
        taxId: values.taxId || undefined,
      })
      .then(() => {
        toast.success("Organization profile updated.");
      })
      .catch((error) => {
        logger.error("Failed to update organization", error);
        toast.error("Something went wrong. Please try again.");
      });
  }

  if (!organization) {
    if (orgListQuery.isError) {
      return <p className="text-sm text-destructive">{"Could not disable the organization."}</p>;
    }
    return <p className="text-sm text-muted-foreground">{"Organization not found."}</p>;
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle>{"Organization profile"}</CardTitle>
            <CardDescription>{"Details that identify your organization."}</CardDescription>
          </div>
          {isActiveOrg ? <Badge variant="secondary">{"Current organization"}</Badge> : null}
        </CardHeader>
        <CardContent>
          <Form form={form} onSubmit={handleSave}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={"Name"} className="sm:col-span-2">
                {({ field }) => <Input {...field} placeholder={"Name"} />}
              </FormField>
              <FormField name="legalName" label={"Legal name"}>
                {({ field }) => <Input {...field} placeholder={"Legal name"} />}
              </FormField>
              <FormField name="parentId" label={"Parent organization"}>
                {({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger aria-label={"Parent organization"}>
                      <SelectValue placeholder={"None (top-level)"} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">{"None (top-level)"}</SelectItem>
                      {parents.map((parent) => (
                        <SelectItem key={parent.id} value={String(parent.id)}>
                          {parent.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="baseCurrency" label={"Base currency"}>
                {({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger aria-label={"Base currency"}>
                      <SelectValue placeholder={"Base currency"} />
                    </SelectTrigger>
                    <SelectContent>
                      {baseCurrencies.map((currency) => (
                        <SelectItem key={currency.code} value={currency.code}>
                          {currency.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="countryCode" label={"Country"}>
                {({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger aria-label={"Country"}>
                      <SelectValue placeholder={"Country"} />
                    </SelectTrigger>
                    <SelectContent>
                      {countries.map((country) => (
                        <SelectItem key={country.code} value={country.code}>
                          {country.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="taxId" label={"Tax ID"}>
                {({ field }) => <Input {...field} placeholder={"Tax ID"} />}
              </FormField>
              <FormField name="timezone" label={"Timezone"}>
                {({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger aria-label={"Timezone"}>
                      <SelectValue placeholder={"Timezone"} />
                    </SelectTrigger>
                    <SelectContent>
                      {timezones.map((timezone) => (
                        <SelectItem key={timezone} value={timezone}>
                          {timezone}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="taxYearStartMonth" label={"Tax year starts in"}>
                {({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger aria-label={"Tax year starts in"}>
                      <SelectValue placeholder={"Tax year starts in"} />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.from({ length: 12 }, (_, index) => index + 1).map((month) => (
                        <SelectItem key={month} value={String(month)}>
                          {new Date(2000, month - 1, 1).toLocaleString("en", { month: "long" })}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
            </div>
            <div className="flex justify-end pt-2">
              <SubmitButton>{"Save changes"}</SubmitButton>
            </div>
          </Form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{"Danger zone"}</CardTitle>
          <CardDescription>{"These actions cannot be undone."}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-start gap-3">
          <p className="text-sm text-muted-foreground">
            {
              "Disabling your organization makes it read-only for all members. You cannot disable the organization you are currently working in."
            }
          </p>
          {disableDisabled ? (
            <Badge variant="outline">
              {"You cannot disable the organization you are currently working in."}
            </Badge>
          ) : null}
          <ConfirmDialog
            title={`Disable ${organization.name}?`}
            description={
              "This will make the organization read-only. Members can still sign in but cannot modify data."
            }
            confirmLabel={"Disable organization"}
            disabled={disableDisabled}
            onConfirm={() => toast.success("Organization disabled.")}
            trigger={
              <Button variant="destructive" disabled={disableDisabled}>
                {"Disable organization"}
              </Button>
            }
          />
        </CardContent>
      </Card>
    </div>
  );
}
