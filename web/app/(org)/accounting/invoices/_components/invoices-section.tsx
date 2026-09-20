"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
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
import type { InvoiceSummary, Journal } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney, getLocalDateString } from "@/lib/utils";
import { type InvoiceState, invoiceStateLabel, invoiceStateTone } from "./invoice-utils";

export function InvoicesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const invoicesQuery = useOrgListQuery<{ invoices: InvoiceSummary[] }, Record<string, never>>(
    "invoices",
    (organizationId) => getSwantaraService().invoices.list(organizationId),
  );
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const invoices = invoicesQuery.data?.invoices ?? [];

  const columns: ColumnDef<InvoiceSummary>[] = [
    {
      accessorKey: "name",
      header: "Number",
      cell: ({ row }) => (
        <a
          href={`/accounting/invoices/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `INV-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "contactId",
      header: "Contact ID",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{`#${row.original.contactId}`}</span>
      ),
    },
    {
      accessorKey: "invoiceDate",
      header: "Invoice date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.invoiceDate ? formatDate(row.original.invoiceDate) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "amountTotal",
      header: "Total",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatMoney(row.original.amountTotal)}</span>
      ),
    },
    {
      accessorKey: "amountResidual",
      header: "Remaining",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatMoney(row.original.amountResidual)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const state = row.original.state as InvoiceState;
        return (
          <Badge variant="outline" className={invoiceStateTone(state)}>
            {invoiceStateLabel(state)}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={invoices}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "posted", label: "Posted" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search invoices..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Accounts"}
        emptyTitle={"No invoices found."}
        status={
          invoicesQuery.isLoading
            ? { type: "loading" }
            : invoicesQuery.isError
              ? {
                  type: "error",
                  message: invoicesQuery.error.message,
                  onRetry: () => void invoicesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New invoice"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <InvoiceFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          journals={journalsQuery.data?.journals ?? []}
          onSave={() => {
            setDialogOpen(false);
            void invoicesQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}

function InvoiceFormDialog({
  open,
  onOpenChange,
  orgId,
  journals,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  journals: Journal[];
  onSave: () => void;
}) {
  const [invoiceType, setInvoiceType] = useState<"customer_invoice" | "vendor_bill">(
    "customer_invoice",
  );
  const [contactId, setContactId] = useState("");
  const [journalId, setJournalId] = useState("");
  const [invoiceDate, setInvoiceDate] = useState(getLocalDateString());
  const [reference, setReference] = useState("");
  const [lines, setLines] = useState<
    { id: string; description: string; qty: string; unitPrice: string }[]
  >([{ id: "1", description: "", qty: "1", unitPrice: "" }]);

  useEffect(() => {
    if (!open) return;
    setInvoiceType("customer_invoice");
    setContactId("");
    setJournalId("");
    setInvoiceDate(getLocalDateString());
    setReference("");
    setLines([{ id: "1", description: "", qty: "1", unitPrice: "" }]);
  }, [open]);

  function addLine() {
    setLines((prev) => [
      ...prev,
      { id: String(Date.now()), description: "", qty: "1", unitPrice: "" },
    ]);
  }

  function updateLine(id: string, patch: Partial<(typeof lines)[number]>) {
    setLines((prev) => prev.map((l) => (l.id === id ? { ...l, ...patch } : l)));
  }

  function removeLine(id: string) {
    setLines((prev) => prev.filter((l) => l.id !== id));
  }

  const subtotal = lines.reduce((s, l) => s + (Number(l.qty) || 0) * (Number(l.unitPrice) || 0), 0);

  function handleSubmit() {
    if (!contactId || !journalId) return;
    void getSwantaraService()
      .invoices.create(Number(orgId) || 0, {
        organizationId: Number(orgId) || 0,
        journalId: Number(journalId) || 0,
        contactId: Number(contactId) || 0,
        date: invoiceDate,
        dueDate: invoiceDate,
        reference,
        lines: lines
          .filter((l) => l.description)
          .map((l) => ({
            itemId: null,
            description: l.description,
            qty: Number(l.qty) || 1,
            unitId: null,
            unitPrice: Number(l.unitPrice) || 0,
            discountPct: 0,
            taxIds: [],
            accountId: 0,
            dimensionId: null,
          })),
      })
      .then(() => {
        toast.success("Invoice created successfully");
        onSave();
      });
  }

  const invoiceTypes = [
    { value: "customer_invoice", label: "Customer invoice" },
    { value: "vendor_bill", label: "Supplier bill" },
  ] as const;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{"Create invoice"}</DialogTitle>
          <DialogDescription>{"Create a new invoice with line items."}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Invoice type"}</span>
              <Select
                value={invoiceType}
                onValueChange={(v) => setInvoiceType(v as "customer_invoice" | "vendor_bill")}
              >
                <SelectTrigger aria-label={"Invoice type"}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {invoiceTypes.map((it) => (
                    <SelectItem key={it.value} value={it.value}>
                      {it.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Contact Id"}</span>
              <Input
                value={contactId}
                onChange={(e) => setContactId(e.target.value)}
                type="number"
                min="1"
              />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Journal"}</span>
              <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
                <SelectTrigger aria-label={"Journal"}>
                  <SelectValue placeholder={"Select a journal"} />
                </SelectTrigger>
                <SelectContent>
                  {journals.map((j) => (
                    <SelectItem key={j.id} value={String(j.id)}>
                      {j.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Invoice date"}</span>
              <Input
                type="date"
                value={invoiceDate}
                onChange={(e) => setInvoiceDate(e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Reference"}</span>
              <Input value={reference} onChange={(e) => setReference(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-3 rounded-md border p-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm">{"Invoice lines"}</h3>
              <Button type="button" variant="outline" size="sm" onClick={addLine}>
                {"Add line"}
              </Button>
            </div>
            {lines.map((line) => (
              <div key={line.id} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
                <div className="sm:col-span-5">
                  <span className="text-xs text-muted-foreground">
                    {"Create a new invoice with line items."}
                  </span>
                  <Input
                    value={line.description}
                    onChange={(e) => updateLine(line.id, { description: e.target.value })}
                  />
                </div>
                <div className="sm:col-span-2">
                  <span className="text-xs text-muted-foreground">{"Quantity"}</span>
                  <Input
                    value={line.qty}
                    onChange={(e) => updateLine(line.id, { qty: e.target.value })}
                    type="number"
                    min="0"
                    step="any"
                  />
                </div>
                <div className="sm:col-span-3">
                  <span className="text-xs text-muted-foreground">{"Unit price"}</span>
                  <Input
                    value={line.unitPrice}
                    onChange={(e) => updateLine(line.id, { unitPrice: e.target.value })}
                    type="number"
                    min="0"
                    step="any"
                  />
                </div>
                <div className="sm:col-span-2 flex items-end">
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => removeLine(line.id)}
                  >
                    {"Remove"}
                  </Button>
                </div>
              </div>
            ))}
            <div className="flex justify-end text-sm">
              <span>
                {"Subtotal"}: {formatMoney(subtotal)}
              </span>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!contactId || !journalId}>
            {"Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
