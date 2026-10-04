"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus, Trash2Icon } from "lucide-react";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Reference");
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
      .min(1, t("validationValueRequired"))
      .refine((value) => Number(value) > 0, t("validationValuePositive")),
    daysAfter: z.string().min(1, t("validationDueDaysRequired")),
    discountPct: z.string(),
    discountDays: z.string(),
  });
  type LineValues = z.infer<typeof lineSchema>;

  const schema = z
    .object({
      name: z.string().min(1, t("validationNameRequired")),
      note: z.string(),
      lines: z.array(lineSchema).min(1, t("validationAddAtLeastOneLine")),
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
          message: t("validationPercentTotal", { total: String(percentTotal) }),
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
      return getSwantaraService()
        .paymentTerms.update(Number(editing.id), payload)
        .then(() => {
          toast.success(t("paymentTermSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update payment term", error);
          toast.error(t("toastFailed"));
        });
    } else {
      return getSwantaraService()
        .paymentTerms.create(payload)
        .then(() => {
          toast.success(t("paymentTermSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create payment term", error);
          toast.error(t("toastFailed"));
        });
    }
  }

  const columns: ColumnDef<PaymentTermRow>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "note",
      header: t("tableNote"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.note || "—"}</span>,
    },
    {
      accessorKey: "lines",
      header: t("schedule"),
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
          editLabel={t("actionEdit")}
          deleteLabel={t("actionDelete")}
          confirmTitle={t("deleteTermTitle")}
          confirmDescription={t("deleteTermDescription")}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .paymentTerms.delete(Number(row.original.id))
              .then(() => void query.refetch())
              .catch((error) => {
                logger.error("Failed to delete payment term", error);
                toast.error(t("toastFailed"));
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
        searchPlaceholder={t("searchTermsPlaceholder")}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={t("paymentTermsTitle")}
        emptyTitle={t("paymentTermsEmpty")}
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
            <span>{t("addTerm")}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("editTerm") : t("newTerm")}</DialogTitle>
            <DialogDescription>{t("paymentTermDialogDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("fieldName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
              </FormField>
              <FormField name="note" label={t("fieldNote")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldNote")} />}
              </FormField>
            </div>

            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <span className="text-sm">{t("schedule")}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => append(emptyLine())}
                >
                  <Plus />
                  <span>{t("addLine")}</span>
                </Button>
              </div>
              {form.formState.errors.lines?.message ? (
                <p role="alert" className="text-sm text-destructive">
                  {form.formState.errors.lines.message}
                </p>
              ) : null}
              <Table aria-label={t("schedule")}>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("tableType")}</TableHead>
                    <TableHead className="w-24">{t("tableValue")}</TableHead>
                    <TableHead className="w-24">{t("tableDueDays")}</TableHead>
                    <TableHead className="w-24">{t("tableDiscountPct")}</TableHead>
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
                          <SelectTrigger aria-label={t("fieldType")} size="sm">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {lineValueTypes.map((type) => (
                              <SelectItem key={type} value={type}>
                                {(t as unknown as (k: string) => string)(`type_${type}`)}
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
                          aria-label={t("fieldValue")}
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
                          aria-label={t("fieldDueDays")}
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
                          aria-label={t("fieldDiscountPct")}
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
                          aria-label={t("removeLine")}
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
                {t("actionCancel")}
              </Button>
              <SubmitButton>{t("saveTerm")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
