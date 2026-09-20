"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/table";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { PaymentTerm } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";
import { logger } from "@/lib/utils/logger";

type LineValueType = "percent" | "fixed" | "balance";

export type PaymentTermLineRow = {
  id: string;
  valueType: LineValueType;
  value: number;
  daysAfter: number;
  discountPct: number;
  discountDays: number;
};

export type PaymentTermRow = {
  id: string;
  name: string;
  note: string;
  lines: PaymentTermLineRow[];
};

function toPaymentTermRow(term: PaymentTerm): PaymentTermRow {
  return {
    id: String(term.id),
    name: term.name,
    note: term.note ?? "",
    lines: [],
  };
}

const lineValueTypes = ["percent", "fixed", "balance"] as const;

export function PaymentTermsSection(_props: { orgId: string }) {
  const [editing, setEditing] = useState<PaymentTermRow | null>(null);
  const [open, setOpen] = useState(false);

  const query = useOrgListQuery<{ paymentTerms: PaymentTerm[] }, Record<string, never>>(
    "paymentTerms",
    () => getSwantaraService().paymentTerms.list(),
  );

  const terms = (query.data?.paymentTerms ?? []).map(toPaymentTermRow);

  const lineSchema = z.object({
    id: z.string(),
    valueType: z.enum(lineValueTypes),
    value: z
      .string()
      .min(1, "Enter a value.")
      .refine((value) => Number(value) > 0, "Value must be positive."),
    daysAfter: z.string().min(1, "Enter due days."),
    discountPct: z.string(),
    discountDays: z.string(),
  });
  type LineValues = z.infer<typeof lineSchema>;

  const schema = z
    .object({
      name: z.string().min(1, "Enter a name."),
      note: z.string(),
      lines: z.array(lineSchema).min(1, "Add at least one line."),
    })
    .superRefine((values, ctx) => {
      const percentTotal = values.lines
        .filter((line) => line.valueType === "percent")
        .reduce((sum, line) => sum + (Number(line.value) || 0), 0);
      if (
        values.lines.some((line) => line.valueType === "percent") &&
        Math.abs(percentTotal - 100) > 0.001
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["lines"],
          message: `Percent lines must add up to 100%. Current total: ${String(percentTotal)}%.`,
        });
      }
    });
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", note: "", lines: [] },
  });
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });
  const lineValues = form.watch("lines");

  function emptyLine(): LineValues {
    return {
      id: `line-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      valueType: "percent",
      value: "",
      daysAfter: "",
      discountPct: "",
      discountDays: "",
    };
  }

  function openCreate() {
    setEditing(null);
    form.reset({ name: "", note: "", lines: [emptyLine()] });
    setOpen(true);
  }

  function openEdit(term: PaymentTermRow) {
    setEditing(term);
    form.reset({
      name: term.name,
      note: term.note,
      lines: term.lines.map((line) => ({
        id: line.id,
        valueType: line.valueType,
        value: String(line.value),
        daysAfter: String(line.daysAfter),
        discountPct: String(line.discountPct),
        discountDays: String(line.discountDays),
      })),
    });
    setOpen(true);
  }

  function handleSubmit(values: Values) {
    const payload = {
      name: values.name,
      note: values.note || undefined,
      lines: values.lines.map((line) => ({
        valueType: line.valueType as "percent" | "fixed" | "balance",
        value: Number(line.value),
        daysAfter: Number(line.daysAfter),
        discountPct: Number(line.discountPct || 0),
        discountDays: Number(line.discountDays || 0),
      })),
    };

    if (editing) {
      void getSwantaraService()
        .paymentTerms.update(Number(editing.id), payload)
        .then(() => {
          toast.success("Payment term saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update payment term", error);
          toast.error("Could not disable the organization.");
        });
    } else {
      void getSwantaraService()
        .paymentTerms.create(payload)
        .then(() => {
          toast.success("Payment term saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create payment term", error);
          toast.error("Could not disable the organization.");
        });
    }
  }

  const columns: ColumnDef<PaymentTermRow>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "note",
      header: "Note",
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.note || "—"}</span>,
    },
    {
      accessorKey: "lines",
      header: "Schedule",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{row.original.lines.length}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this term?"}
          confirmDescription={"The payment term will no longer be available on documents."}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .paymentTerms.delete(Number(row.original.id))
              .then(() => void query.refetch())
              .catch((error) => {
                logger.error("Failed to delete payment term", error);
                toast.error("Could not disable the organization.");
              })
          }
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={terms}
        getRowId={(row) => row.id}
        searchKeys={["name", "note"]}
        searchPlaceholder={"Search terms…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"Payment terms"}
        emptyTitle={"No payment terms"}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={openCreate}>
            <Plus />
            <span>{"Add term"}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit payment term" : "New payment term"}</DialogTitle>
            <DialogDescription>
              {"Terms applied to customer invoices and supplier bills."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="note" label={"Note"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Note"} />}
              </FormField>
            </div>

            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <span className="text-sm">{"Schedule"}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => append(emptyLine())}
                >
                  <Plus />
                  <span>{"Add line"}</span>
                </Button>
              </div>
              {form.formState.errors.lines?.message ? (
                <p role="alert" className="text-sm text-destructive">
                  {form.formState.errors.lines.message}
                </p>
              ) : null}
              <Table aria-label={"Schedule"}>
                <TableHeader>
                  <TableRow>
                    <TableHead>{"Type"}</TableHead>
                    <TableHead className="w-24">{"Value"}</TableHead>
                    <TableHead className="w-24">{"Due days"}</TableHead>
                    <TableHead className="w-24">{"Discount %"}</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {fields.map((line, index) => (
                    <TableRow key={line.id}>
                      <TableCell>
                        <Select
                          value={lineValues[index]!.valueType}
                          onValueChange={(value) =>
                            form.setValue(
                              `lines.${index}.valueType`,
                              (value ?? "percent") as LineValueType,
                            )
                          }
                        >
                          <SelectTrigger aria-label={"Type"} size="sm">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {lineValueTypes.map((type) => (
                              <SelectItem key={type} value={type}>
                                {humanizeKey(String(type))}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="any"
                          inputMode="decimal"
                          aria-label={"Value"}
                          value={lineValues[index]!.value}
                          onChange={(event) =>
                            form.setValue(`lines.${index}.value`, event.target.value)
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="any"
                          inputMode="decimal"
                          aria-label={"Due days"}
                          value={lineValues[index]!.daysAfter}
                          onChange={(event) =>
                            form.setValue(`lines.${index}.daysAfter`, event.target.value)
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="any"
                          inputMode="decimal"
                          aria-label={"Discount %"}
                          value={lineValues[index]!.discountPct}
                          onChange={(event) =>
                            form.setValue(`lines.${index}.discountPct`, event.target.value)
                          }
                        />
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={"Remove line"}
                          onClick={() => remove(index)}
                          className="text-destructive"
                        >
                          <Trash2Icon />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save term"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
