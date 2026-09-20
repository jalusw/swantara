"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
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
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { FxRate } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { logger } from "@/lib/utils/logger";

type RateType = "spot" | "avg" | "closing";

export type FxRateRow = {
  id: string;
  currencyCode: string;
  rate: number;
  rateType: RateType;
  validFrom: string;
};

function toFxRateRow(rate: FxRate): FxRateRow {
  return {
    id: String(rate.id),
    currencyCode: rate.currencyCode,
    rate: rate.rate,
    rateType: rate.rateType,
    validFrom:
      rate.validFrom instanceof Date
        ? rate.validFrom.toISOString().split("T")[0]!
        : String(rate.validFrom),
  };
}

const currencyOptions = ["USD", "IDR", "EUR", "SGD", "GBP", "JPY"];

const rateTypes: RateType[] = ["spot", "avg", "closing"];

function rateTypeLabel(rateType: RateType): string {
  if (rateType === "avg") return "Average";
  return humanizeKey(rateType);
}

function useFxRateSchema() {
  return z.object({
    currencyCode: z.string().min(1, "Select a currency."),
    rate: z
      .string()
      .min(1, "Enter a rate.")
      .refine((value) => Number(value) > 0, "Rate must be greater than zero."),
    rateType: z.enum(["spot", "avg", "closing"]),
    validFrom: z.string().min(1, "Enter a valid date."),
  });
}
type FxRateValues = z.infer<ReturnType<typeof useFxRateSchema>>;

export function FxRatesSection({ orgId }: { orgId: string }) {
  const [editing, setEditing] = useState<FxRateRow | null>(null);
  const [open, setOpen] = useState(false);

  const query = useOrgListQuery<{ fxRates: FxRate[] }, Record<string, never>>(
    "fxRates",
    (organizationId) => getSwantaraService().fxRates.list(organizationId),
  );

  const rates = (query.data?.fxRates ?? []).map(toFxRateRow);

  const schema = useFxRateSchema();

  const form = useForm<FxRateValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      currencyCode: "",
      rate: "",
      rateType: "spot",
      validFrom: "",
    },
  });

  function openCreate() {
    setEditing(null);
    form.reset({ currencyCode: "", rate: "", rateType: "spot", validFrom: "" });
    setOpen(true);
  }

  function openEdit(row: FxRateRow) {
    setEditing(row);
    form.reset({
      currencyCode: row.currencyCode,
      rate: String(row.rate),
      rateType: row.rateType,
      validFrom: row.validFrom,
    });
    setOpen(true);
  }

  function handleSubmit(values: FxRateValues) {
    const payload = {
      currencyCode: values.currencyCode,
      rate: Number(values.rate),
      rateType: values.rateType as RateType,
      validFrom: values.validFrom,
    };

    if (editing) {
      void getSwantaraService()
        .fxRates.update(Number(orgId), Number(editing.id), payload)
        .then(() => {
          toast.success("Exchange rate saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update fx rate", error);
          toast.error("Could not disable the organization.");
        });
    } else {
      void getSwantaraService()
        .fxRates.create(Number(orgId), payload)
        .then(() => {
          toast.success("Exchange rate saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create fx rate", error);
          toast.error("Could not disable the organization.");
        });
    }
  }

  const columns: ColumnDef<FxRateRow>[] = [
    {
      accessorKey: "currencyCode",
      header: "Currency",
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.currencyCode}</span>,
    },
    {
      accessorKey: "rate",
      header: "Rate",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.rate, { maximumFractionDigits: 6 })}
        </span>
      ),
    },
    {
      accessorKey: "rateType",
      header: "Rate type",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{rateTypeLabel(row.original.rateType)}</span>
      ),
    },
    {
      accessorKey: "validFrom",
      header: "Valid from",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{formatDate(row.original.validFrom)}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this rate?"}
          confirmDescription={"The rate will be removed and can no longer be used for conversions."}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .fxRates.delete(Number(orgId), Number(row.original.id))
              .then(() => void query.refetch())
              .catch((error) => {
                logger.error("Failed to delete fx rate", error);
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
        data={rates}
        getRowId={(row) => row.id}
        searchKeys={["currencyCode"]}
        searchPlaceholder={"Search rates…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"FX rates"}
        emptyTitle={"No exchange rates"}
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
            <span>{"Add rate"}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? "Edit exchange rate" : "New exchange rate"}</DialogTitle>
            <DialogDescription>
              {"Exchange rates used to convert amounts between currencies."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="currencyCode" label={"Currency"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Currency"}>
                      <SelectValue placeholder={"All currencies"} />
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
              <FormField name="rateType" label={"Rate type"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Rate type"}>
                      <SelectValue placeholder={"Rate type"} />
                    </SelectTrigger>
                    <SelectContent>
                      {rateTypes.map((type) => (
                        <SelectItem key={type} value={type}>
                          {rateTypeLabel(type)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="rate" label={"Rate"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="0.000001" inputMode="decimal" />
                )}
              </FormField>
              <FormField name="validFrom" label={"Valid from"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save rate"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
