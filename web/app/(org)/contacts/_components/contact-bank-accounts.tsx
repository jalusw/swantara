"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Landmark, Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
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
import { useContactBankAccountsQuery } from "@/lib/hooks/use-contact-query";
import type { ContactBankAccount } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const currencyOptions = ["USD", "IDR", "EUR", "SGD", "GBP", "JPY"];

function useContactBankAccountFormSchema() {
  return z.object({
    accountHolder: z.string().min(1, "Enter the account holder."),
    bankName: z.string(),
    iban: z.string(),
    swiftBic: z.string(),
    accountNumber: z.string(),
    routingNumber: z.string(),
    currencyCode: z.string(),
  });
}
type ContactBankAccountFormValues = z.infer<ReturnType<typeof useContactBankAccountFormSchema>>;

export function ContactBankAccounts({
  orgId,
  contactId,
  onRefetch,
}: {
  orgId: string;
  contactId: string;
  onRefetch: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ContactBankAccount | null>(null);

  const bankAccountsQuery = useContactBankAccountsQuery(orgId, contactId);
  const bankAccounts = bankAccountsQuery.data ?? [];

  const schema = useContactBankAccountFormSchema();

  const form = useForm<ContactBankAccountFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      accountHolder: "",
      bankName: "",
      iban: "",
      swiftBic: "",
      accountNumber: "",
      routingNumber: "",
      currencyCode: "USD",
    },
  });

  function openCreate() {
    setEditing(null);
    form.reset({
      accountHolder: "",
      bankName: "",
      iban: "",
      swiftBic: "",
      accountNumber: "",
      routingNumber: "",
      currencyCode: "USD",
    });
    setOpen(true);
  }

  function openEdit(account: ContactBankAccount) {
    setEditing(account);
    form.reset({
      accountHolder: account.accountHolder ?? "",
      bankName: account.bankName ?? "",
      iban: account.iban ?? "",
      swiftBic: account.swiftBic ?? "",
      accountNumber: account.accountNumber ?? "",
      routingNumber: account.routingNumber ?? "",
      currencyCode: account.currencyCode ?? "USD",
    });
    setOpen(true);
  }

  function handleSubmit(values: ContactBankAccountFormValues) {
    const request = {
      accountHolder: values.accountHolder || null,
      bankName: values.bankName || null,
      iban: values.iban || null,
      swiftBic: values.swiftBic || null,
      accountNumber: values.accountNumber || null,
      routingNumber: values.routingNumber || null,
      currencyCode: values.currencyCode || null,
    };

    if (editing) {
      void getSwantaraService()
        .contacts.bankAccounts.update(Number(orgId), Number(contactId), editing.id, request)
        .then(() => {
          toast.success("Bank account saved.");
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error("Something went wrong. Please try again.");
        });
    } else {
      void getSwantaraService()
        .contacts.bankAccounts.create(Number(orgId), Number(contactId), request)
        .then(() => {
          toast.success("Bank account saved.");
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error("Something went wrong. Please try again.");
        });
    }
  }

  function deleteAccount(account: ContactBankAccount) {
    void getSwantaraService()
      .contacts.bankAccounts.delete(Number(orgId), Number(contactId), account.id)
      .then(() => {
        toast.success("Bank account saved.");
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
          <CardTitle>{"Bank accounts"}</CardTitle>
          <CardDescription>{"Accounts used to receive or make payments."}</CardDescription>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{"Add bank account"}</span>
        </Button>
      </CardHeader>
      <CardContent>
        {bankAccounts.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-10 text-center">
            <Landmark className="size-6 text-muted-foreground" aria-hidden />
            <p className="text-sm text-muted-foreground">{"No bank accounts"}</p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {bankAccounts.map((account) => (
              <div
                key={account.id}
                className="flex items-start justify-between gap-4 rounded-md border p-4"
              >
                <div className="flex flex-col gap-1">
                  <p className="text-sm">{account.accountHolder}</p>
                  <p className="text-sm text-muted-foreground">
                    {account.bankName}
                    {account.bankName && account.currencyCode ? " · " : ""}
                    {account.currencyCode}
                  </p>
                  {account.accountNumber ? (
                    <p className="font-mono text-sm text-muted-foreground">
                      {account.accountNumber}
                    </p>
                  ) : null}
                  {account.swiftBic ? (
                    <p className="font-mono text-xs text-muted-foreground">{account.swiftBic}</p>
                  ) : null}
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  <Button variant="ghost" size="sm" onClick={() => openEdit(account)}>
                    {"Edit"}
                  </Button>
                  <ConfirmDialog
                    title={"Delete this bank account?"}
                    description={"The bank account will be removed from the contact."}
                    confirmLabel={"Delete"}
                    cancelLabel={"Cancel"}
                    onConfirm={() => deleteAccount(account)}
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
            <DialogTitle>{editing ? "Edit bank account" : "New bank account"}</DialogTitle>
            <DialogDescription>{"Accounts used to receive or make payments."}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="accountHolder" label={"Account holder"} className="sm:col-span-2">
                {({ field, id }) => <Input {...field} id={id} placeholder={"Account holder"} />}
              </FormField>
              <FormField name="bankName" label={"Bank name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Bank name"} />}
              </FormField>
              <FormField name="currencyCode" label={"Currency"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Currency"}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {currencyOptions.map((code) => (
                        <SelectItem key={code} value={code}>
                          {code}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="accountNumber" label={"Account number"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Account number"} />}
              </FormField>
              <FormField name="iban" label={"IBAN"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"IBAN"} />}
              </FormField>
              <FormField name="swiftBic" label={"SWIFT / BIC"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"SWIFT / BIC"} />}
              </FormField>
              <FormField name="routingNumber" label={"Routing number"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Routing number"} />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save bank account"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
