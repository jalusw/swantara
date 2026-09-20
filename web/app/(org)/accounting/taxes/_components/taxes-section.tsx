"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
import { formatTaxAmount, taxScopeLabels, taxTypeLabels, taxTypeTone } from "./tax-utils";

export function TaxesSection({ orgId }: { orgId: string }) {
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
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <Badge variant="outline" className={taxTypeTone(row.original.type)}>
          {taxTypeLabels[row.original.type]}
        </Badge>
      ),
    },
    {
      accessorKey: "amount",
      header: "Amount",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatTaxAmount(row.original)}</span>,
    },
    {
      accessorKey: "scope",
      header: "Scope",
      cell: ({ row }) => <Badge variant="secondary">{taxScopeLabels[row.original.scope]}</Badge>,
    },
    {
      accessorKey: "priceInclude",
      header: "Included in price",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.priceInclude ? "✓" : "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete tax"}
          confirmDescription={"Are you sure you want to delete this tax?"}
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
        searchPlaceholder={"Search taxes..."}
        emptyTitle={"No taxes configured."}
        ariaLabel={"Accounts"}
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
            <span>{"Add tax"}</span>
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
      toast.success(tax ? "Tax updated successfully" : "Tax created successfully");
      onSave();
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{tax ? "Edit tax" : "Create tax"}</DialogTitle>
          <DialogDescription>{"Define the tax name, rate, and scope."}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Name"}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Amount"}</span>
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
              <span className="text-sm">{"Type"}</span>
              <Select value={type} onValueChange={(v) => setType(v as Tax["type"])}>
                <SelectTrigger aria-label={"Type"}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {taxTypes.map((tp) => (
                    <SelectItem key={tp} value={tp}>
                      {taxTypeLabels[tp]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Scope"}</span>
              <Select value={scope} onValueChange={(v) => setScope(v as Tax["scope"])}>
                <SelectTrigger aria-label={"Scope"}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {taxScopes.map((sc) => (
                    <SelectItem key={sc} value={sc}>
                      {taxScopeLabels[sc]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Tax account"}</span>
            <Select value={taxAccountId} onValueChange={(v) => setTaxAccountId(v ?? "")}>
              <SelectTrigger aria-label={"Tax account"}>
                <SelectValue placeholder={"None"} />
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
            {"Included in price"}
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!name}>
            {"Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
