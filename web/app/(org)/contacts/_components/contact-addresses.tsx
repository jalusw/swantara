"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { MapPin, Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { ConfirmDialog } from "@/components/confirm-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Switch } from "@/components/switch";
import { useContactAddressesQuery } from "@/lib/hooks/use-contact-query";
import type { ContactAddress } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";

const addressTypes = ["billing", "shipping", "other"] as const;
const countryCodeOptions = [
  { code: "US", name: "United States" },
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapore" },
  { code: "DE", name: "Germany" },
  { code: "GB", name: "United Kingdom" },
  { code: "JP", name: "Japan" },
  { code: "AU", name: "Australia" },
];

function useContactAddressFormSchema() {
  return z.object({
    type: z.enum(addressTypes),
    line1: z.string().min(1, "Enter the first address line."),
    line2: z.string(),
    city: z.string(),
    state: z.string(),
    postalCode: z.string(),
    countryCode: z.string(),
    isDefault: z.boolean(),
  });
}
type ContactAddressFormValues = z.infer<ReturnType<typeof useContactAddressFormSchema>>;

export function ContactAddresses({
  orgId,
  contactId,
  onRefetch,
}: {
  orgId: string;
  contactId: string;
  onRefetch: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ContactAddress | null>(null);

  const addressesQuery = useContactAddressesQuery(orgId, contactId);
  const addresses = addressesQuery.data ?? [];

  const schema = useContactAddressFormSchema();

  const form = useForm<ContactAddressFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: "billing",
      line1: "",
      line2: "",
      city: "",
      state: "",
      postalCode: "",
      countryCode: "US",
      isDefault: false,
    },
  });

  function openCreate() {
    setEditing(null);
    form.reset({
      type: "billing",
      line1: "",
      line2: "",
      city: "",
      state: "",
      postalCode: "",
      countryCode: "US",
      isDefault: addresses.length === 0,
    });
    setOpen(true);
  }

  function openEdit(address: ContactAddress) {
    setEditing(address);
    form.reset({
      type: address.type ?? "billing",
      line1: address.line1 ?? "",
      line2: address.line2 ?? "",
      city: address.city ?? "",
      state: address.state ?? "",
      postalCode: address.postalCode ?? "",
      countryCode: address.countryCode ?? "US",
      isDefault: address.isDefault,
    });
    setOpen(true);
  }

  function handleSubmit(values: ContactAddressFormValues) {
    const request = {
      type: values.type,
      line1: values.line1 || null,
      line2: values.line2 || null,
      city: values.city || null,
      state: values.state || null,
      postalCode: values.postalCode || null,
      countryCode: values.countryCode || null,
      isDefault: values.isDefault,
    };

    if (editing) {
      void getSwantaraService()
        .contacts.addresses.update(Number(orgId), Number(contactId), editing.id, request)
        .then(() => {
          toast.success("Address saved.");
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error("Something went wrong. Please try again.");
        });
    } else {
      void getSwantaraService()
        .contacts.addresses.create(Number(orgId), Number(contactId), request)
        .then(() => {
          toast.success("Address saved.");
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error("Something went wrong. Please try again.");
        });
    }
  }

  function makeDefault(address: ContactAddress) {
    void getSwantaraService()
      .contacts.addresses.setDefault(Number(orgId), Number(contactId), address.id)
      .then(() => {
        toast.success("Default address updated.");
        onRefetch();
      })
      .catch(() => {
        toast.error("Something went wrong. Please try again.");
      });
  }

  function deleteAddress(address: ContactAddress) {
    void getSwantaraService()
      .contacts.addresses.delete(Number(orgId), Number(contactId), address.id)
      .then(() => {
        toast.success("Address saved.");
        onRefetch();
      })
      .catch(() => {
        toast.error("Something went wrong. Please try again.");
      });
  }

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between">
        <div className="flex flex-col gap-1">
          <CardTitle>{"Addresses"}</CardTitle>
          <CardDescription>
            {"Contact addresses. One address can be the default for the contact."}
          </CardDescription>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{"Add address"}</span>
        </Button>
      </CardHeader>
      <CardContent>
        {addresses.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-10 text-center">
            <MapPin className="size-6 text-muted-foreground" aria-hidden />
            <p className="text-sm text-muted-foreground">{"No addresses"}</p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {addresses.map((address) => (
              <div
                key={address.id}
                className="flex items-start justify-between gap-4 rounded-md border p-4"
              >
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary">
                      {humanizeKey(String(address.type ?? "other"))}
                    </Badge>
                    {address.isDefault ? <Badge variant="default">{"Default"}</Badge> : null}
                  </div>
                  {address.line1 ? <p className="text-sm">{address.line1}</p> : null}
                  {address.line2 ? (
                    <p className="text-sm text-muted-foreground">{address.line2}</p>
                  ) : null}
                  <p className="text-sm text-muted-foreground">
                    {[address.city, address.state, address.postalCode].filter(Boolean).join(", ")}
                    {address.countryCode ? ` · ${address.countryCode}` : ""}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  {!address.isDefault ? (
                    <Button variant="ghost" size="sm" onClick={() => makeDefault(address)}>
                      {"Set as default"}
                    </Button>
                  ) : null}
                  <Button variant="ghost" size="sm" onClick={() => openEdit(address)}>
                    {"Edit"}
                  </Button>
                  <ConfirmDialog
                    title={"Delete this address?"}
                    description={"The address will be removed from the contact."}
                    confirmLabel={"Delete"}
                    cancelLabel={"Cancel"}
                    onConfirm={() => deleteAddress(address)}
                    trigger={
                      <Button variant="ghost" size="sm" className="text-destructive">
                        {"Delete"}
                      </Button>
                    }
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? "Edit address" : "New address"}</DialogTitle>
            <DialogDescription>
              {"Contact addresses. One address can be the default for the contact."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="type" label={"Type"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Type"}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {addressTypes.map((type) => (
                        <SelectItem key={type} value={type}>
                          {humanizeKey(String(type))}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="countryCode" label={"Country"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Country"}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {countryCodeOptions.map((country) => (
                        <SelectItem key={country.code} value={country.code}>
                          {country.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="line1" label={"Address line 1"} className="sm:col-span-2">
                {({ field, id }) => <Input {...field} id={id} placeholder={"Address line 1"} />}
              </FormField>
              <FormField name="line2" label={"Address line 2"} className="sm:col-span-2">
                {({ field, id }) => <Input {...field} id={id} placeholder={"Address line 2"} />}
              </FormField>
              <FormField name="city" label={"City"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"City"} />}
              </FormField>
              <FormField name="state" label={"State / region"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"State / region"} />}
              </FormField>
              <FormField name="postalCode" label={"Postal code"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Postal code"} />}
              </FormField>
              <FormField name="isDefault" label={"Default"}>
                {({ field, id }) => (
                  <Switch
                    id={id}
                    checked={field.value}
                    onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                    aria-label={"Default"}
                  />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save address"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
