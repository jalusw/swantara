"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Landmark, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Contacts");
  return z.object({
    accountHolder: z.string().min(1, t("validationAccountHolderRequired")),
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
  const t = useTranslations("Contacts");

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
      return getSwantaraService()
        .contacts.bankAccounts.update(Number(orgId), Number(contactId), editing.id, request)
        .then(() => {
          toast.success(t("bankAccountSaved"));
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error(t("saveFailedTryAgain"));
        });
    } else {
      return getSwantaraService()
        .contacts.bankAccounts.create(Number(orgId), Number(contactId), request)
        .then(() => {
          toast.success(t("bankAccountSaved"));
          setOpen(false);
          onRefetch();
        })
        .catch(() => {
          toast.error(t("saveFailedTryAgain"));
        });
    }
  }

  function deleteAccount(account: ContactBankAccount) {
    void getSwantaraService()
      .contacts.bankAccounts.delete(Number(orgId), Number(contactId), account.id)
      .then(() => {
        toast.success(t("bankAccountSaved"));
        onRefetch();
      })
      .catch(() => {
        toast.error(t("saveFailedTryAgain"));
      });
  }

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between">
        <div className="flex flex-col gap-1">
          <CardTitle>{t("bankAccountsTitle")}</CardTitle>
          <CardDescription>{t("bankAccountsDescription")}</CardDescription>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{t("addBankAccount")}</span>
        </Button>
      </CardHeader>
      <CardContent>
        {bankAccounts.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-10 text-center">
            <Landmark className="size-6 text-muted-foreground" aria-hidden />
            <p className="text-sm text-muted-foreground">{t("noBankAccounts")}</p>
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
                    {t("edit")}
                  </Button>
                  <ConfirmDialog
                    title={t("deleteBankAccountTitle")}
                    description={t("deleteBankAccountDescription")}
                    confirmLabel={t("delete")}
                    cancelLabel={t("cancel")}
                    onConfirm={() => deleteAccount(account)}
                    trigger={
                      <Button variant="ghost" size="sm" className="text-destructive">
                        {t("delete")}
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
            <DialogTitle>{editing ? t("editBankAccount") : t("newBankAccount")}</DialogTitle>
            <DialogDescription>{t("bankAccountsDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField
                name="accountHolder"
                label={t("fieldAccountHolder")}
                className="sm:col-span-2"
              >
                {({ field, id }) => (
                  <Input {...field} id={id} placeholder={t("fieldAccountHolder")} />
                )}
              </FormField>
              <FormField name="bankName" label={t("fieldBankName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldBankName")} />}
              </FormField>
              <FormField name="currencyCode" label={t("fieldCurrency")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldCurrency")}>
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
              <FormField name="accountNumber" label={t("fieldAccountNumber")}>
                {({ field, id }) => (
                  <Input {...field} id={id} placeholder={t("fieldAccountNumber")} />
                )}
              </FormField>
              <FormField name="iban" label={t("fieldIban")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldIban")} />}
              </FormField>
              <FormField name="swiftBic" label={t("fieldSwiftBic")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldSwiftBic")} />}
              </FormField>
              <FormField name="routingNumber" label={t("fieldRoutingNumber")}>
                {({ field, id }) => (
                  <Input {...field} id={id} placeholder={t("fieldRoutingNumber")} />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {t("cancel")}
              </Button>
              <SubmitButton>{t("saveBankAccount")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
