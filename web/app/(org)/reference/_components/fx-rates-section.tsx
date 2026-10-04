"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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

function useFxRateSchema() {
  const t = useTranslations("Reference");
  return z.object({
    currencyCode: z.string().min(1, t("validationCurrencyRequired")),
    rate: z
      .string()
      .min(1, t("validationRateRequired"))
      .refine((value) => Number(value) > 0, t("validationRatePositive")),
    rateType: z.enum(["spot", "avg", "closing"]),
    validFrom: z.string().min(1, t("validationDateRequired")),
  });
}
type FxRateValues = z.infer<ReturnType<typeof useFxRateSchema>>;

export function FxRatesSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Reference");
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
      return getSwantaraService()
        .fxRates.update(Number(orgId), Number(editing.id), payload)
        .then(() => {
          toast.success(t("fxRateSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update fx rate", error);
          toast.error(t("toastFailed"));
        });
    } else {
      return getSwantaraService()
        .fxRates.create(Number(orgId), payload)
        .then(() => {
          toast.success(t("fxRateSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create fx rate", error);
          toast.error(t("toastFailed"));
        });
    }
  }

  const columns: ColumnDef<FxRateRow>[] = [
    {
      accessorKey: "currencyCode",
      header: t("tableCurrency"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.currencyCode}</span>,
    },
    {
      accessorKey: "rate",
      header: t("tableRate"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.rate, { maximumFractionDigits: 6 })}
        </span>
      ),
    },
    {
      accessorKey: "rateType",
      header: t("tableRateType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {(t as unknown as (k: string) => string)(`rateType_${row.original.rateType}`)}
        </span>
      ),
    },
    {
      accessorKey: "validFrom",
      header: t("tableValidFrom"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{formatDate(row.original.validFrom)}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("actionEdit")}
          deleteLabel={t("actionDelete")}
          confirmTitle={t("deleteRateTitle")}
          confirmDescription={t("deleteRateDescription")}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .fxRates.delete(Number(orgId), Number(row.original.id))
              .then(() => void query.refetch())
              .catch((error) => {
                logger.error("Failed to delete fx rate", error);
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
        data={rates}
        getRowId={(row) => row.id}
        searchKeys={["currencyCode"]}
        searchPlaceholder={t("searchRatesPlaceholder")}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={t("fxRatesTitle")}
        emptyTitle={t("fxRatesEmpty")}
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
            <span>{t("addRate")}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t("editFxRate") : t("newFxRate")}</DialogTitle>
            <DialogDescription>{t("fxRateDialogDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="currencyCode" label={t("fieldCurrency")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldCurrency")}>
                      <SelectValue placeholder={t("allCurrencies")} />
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
              <FormField name="rateType" label={t("fieldRateType")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldRateType")}>
                      <SelectValue placeholder={t("fieldRateType")} />
                    </SelectTrigger>
                    <SelectContent>
                      {rateTypes.map((type) => (
                        <SelectItem key={type} value={type}>
                          {(t as unknown as (k: string) => string)(`rateType_${type}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="rate" label={t("fieldRate")}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="0.000001" inputMode="decimal" />
                )}
              </FormField>
              <FormField name="validFrom" label={t("fieldValidFrom")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {t("actionCancel")}
              </Button>
              <SubmitButton>{t("saveRate")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
