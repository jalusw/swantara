"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Account, Tax } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatTaxAmount, taxTypeTone } from "./tax-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function TaxesSection({ orgId }: { orgId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTax, setEditingTax] = useState<Tax | null>(null);

  const taxesQuery = useOrgListQuery<{ taxes: Tax[] }, Record<string, never>>(
    "taxes",
    (organizationId) => getSwantaraService().taxes.list(organizationId),
  );
  const accountsQuery = useOrgListQuery<{ accounts: Account[] }, Record<string, never>>(
    "accounts",
    (organizationId) => getSwantaraService().accounts.list(organizationId),
  );

  const taxes = taxesQuery.data?.taxes ?? [];
  const accounts = accountsQuery.data?.accounts ?? [];

  function handleEdit(tax: Tax) {
    setEditingTax(tax);
    setDialogOpen(true);
  }

  function handleCreate() {
    setEditingTax(null);
    setDialogOpen(true);
  }

  function handleDelete(tax: Tax) {
    void getSwantaraService()
      .taxes.delete(Number(orgId), tax.id)
      .then(() => void taxesQuery.refetch());
  }

  function handleSaved() {
    setDialogOpen(false);
    setEditingTax(null);
    void taxesQuery.refetch();
  }

  const columns: ColumnDef<Tax>[] = [
    {
      accessorKey: "name",
      header: () => t("fieldName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "type",
      header: () => t("fieldType"),
      cell: ({ row }) => (
        <Badge variant="outline" className={taxTypeTone(row.original.type)}>
          {(t as unknown as (k: string) => string)(`taxType_${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "amount",
      header: () => t("colAmount"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatTaxAmount(row.original)}</span>,
    },
    {
      accessorKey: "scope",
      header: () => t("fieldScope"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {(t as unknown as (k: string) => string)(`taxScope_${row.original.scope}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "priceInclude",
      header: () => t("colPriceInclude"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.priceInclude ? "✓" : "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteTax")}
          confirmDescription={t("deleteTaxDescription")}
          onEdit={() => handleEdit(row.original)}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={taxes}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={t("searchTaxes")}
        emptyTitle={t("taxesEmpty")}
        ariaLabel={t("taxesTitle")}
        status={
          taxesQuery.isLoading
            ? { type: "loading" }
            : taxesQuery.isError
              ? {
                  type: "error",
                  message: taxesQuery.error.message,
                  onRetry: () => void taxesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={handleCreate}>
            <Plus />
            <span>{t("addTax")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <TaxFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          tax={editingTax}
          accounts={accounts}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function TaxFormDialog({
  open,
  onOpenChange,
  orgId,
  tax,
  accounts,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  tax: Tax | null;
  accounts: Account[];
  onSave: () => void;
}) {
  const [name, setName] = useState(tax?.name ?? "");
  const [amount, setAmount] = useState(String(tax?.amount ?? ""));
  const [type, setType] = useState<Tax["type"]>(tax?.type ?? "percent");
  const [scope, setScope] = useState<Tax["scope"]>(tax?.scope ?? "sale");
  const [priceInclude, setPriceInclude] = useState(tax?.priceInclude ?? false);
  const [taxAccountId, setTaxAccountId] = useState(
    tax?.taxAccountId != null ? String(tax.taxAccountId) : "",
  );
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");

  const taxTypes = ["percent", "fixed", "group"] as const;
  const taxScopes = ["sale", "purchase", "none"] as const;

  function handleSubmit() {
    if (!name) return;
    const request = {
      name,
      amount: amount ? Number(amount) : null,
      type,
      scope,
      priceInclude,
      taxAccountId: taxAccountId ? Number(taxAccountId) : null,
      refundTaxAccountId: null as number | null,
    };
    const promise = tax
      ? getSwantaraService().taxes.update(Number(orgId), tax.id, request)
      : getSwantaraService().taxes.create(Number(orgId), request);
    void promise.then(() => {
      toast.success(tax ? t("toastTaxUpdated") : t("toastTaxCreated"));
      onSave();
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{tax ? t("editTax") : t("createTax")}</DialogTitle>
          <DialogDescription>{t("taxDialogDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldName")}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("colAmount")}</span>
              <Input
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                type="number"
                min="0"
                step="any"
                disabled={type === "group"}
              />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldType")}</span>
              <Select value={type} onValueChange={(v) => setType(v as Tax["type"])}>
                <SelectTrigger aria-label={t("fieldType")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {taxTypes.map((tp) => (
                    <SelectItem key={tp} value={tp}>
                      {(t as unknown as (k: string) => string)(`taxType_${tp}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldScope")}</span>
              <Select value={scope} onValueChange={(v) => setScope(v as Tax["scope"])}>
                <SelectTrigger aria-label={t("fieldScope")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {taxScopes.map((sc) => (
                    <SelectItem key={sc} value={sc}>
                      {(t as unknown as (k: string) => string)(`taxScope_${sc}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldTaxAccount")}</span>
            <Select value={taxAccountId} onValueChange={(v) => setTaxAccountId(v ?? "")}>
              <SelectTrigger aria-label={t("fieldTaxAccount")}>
                <SelectValue placeholder={t("noneLabel")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} — {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={priceInclude}
              onChange={(e) => setPriceInclude(e.target.checked)}
              className="size-4"
            />
            {t("colPriceInclude")}
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!name}>
            {tCommon("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
