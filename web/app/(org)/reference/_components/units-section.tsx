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
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
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
import type { Unit, UnitCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { logger } from "@/lib/utils/logger";

export type UnitCategoryRow = {
  id: string;
  name: string;
};

export type UnitRow = {
  id: string;
  categoryId: string;
  name: string;
  factor: number;
  uomType: "reference" | "bigger" | "smaller";
  rounding: string;
};

function toCategoryRow(category: UnitCategory): UnitCategoryRow {
  return { id: String(category.id), name: category.name };
}

function toUnitRow(unit: Unit): UnitRow {
  return {
    id: String(unit.id),
    categoryId: String(unit.categoryId),
    name: unit.name,
    factor: unit.factor,
    uomType: (unit.uomType as UnitRow["uomType"]) ?? "reference",
    rounding: String(unit.rounding),
  };
}

const uomTypes = ["reference", "bigger", "smaller"] as const;

function useUnitSchema() {
  const t = useTranslations("Reference");
  return z.object({
    categoryId: z.string().min(1, t("validationCategoryRequired")),
    name: z.string().min(1, t("validationNameRequired")),
    factor: z
      .string()
      .min(1, t("validationFactorRequired"))
      .refine((value) => Number(value) > 0, t("validationFactorPositive")),
    uomType: z.enum(uomTypes),
    rounding: z.string(),
  });
}
type UnitValues = z.infer<ReturnType<typeof useUnitSchema>>;

function useUnitCategorySchema() {
  const t = useTranslations("Reference");
  return z.object({
    name: z.string().min(1, t("validationCategoryNameRequired")),
  });
}
type UnitCategoryValues = z.infer<ReturnType<typeof useUnitCategorySchema>>;

export function UnitsSection(_props: { orgId: string }) {
  const t = useTranslations("Reference");
  const [editing, setEditing] = useState<UnitRow | null>(null);
  const [open, setOpen] = useState(false);
  const [categoryOpen, setCategoryOpen] = useState(false);

  const categoriesQuery = useOrgListQuery<{ categories: UnitCategory[] }, Record<string, never>>(
    "uomCategories",
    () => getSwantaraService().uomCategories.list(),
  );

  const unitsQuery = useOrgListQuery<{ units: Unit[] }, Record<string, never>>("units", () =>
    getSwantaraService().units.list(),
  );

  const categories = (categoriesQuery.data?.categories ?? []).map(toCategoryRow);
  const units = (unitsQuery.data?.units ?? []).map(toUnitRow);

  const schema = useUnitSchema();

  const form = useForm<UnitValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      categoryId: "",
      name: "",
      factor: "",
      uomType: "reference",
      rounding: "0.001",
    },
  });

  const categorySchema = useUnitCategorySchema();

  const categoryForm = useForm<UnitCategoryValues>({
    resolver: zodResolver(categorySchema),
    defaultValues: { name: "" },
  });

  function openCreate() {
    setEditing(null);
    form.reset({
      categoryId: "",
      name: "",
      factor: "",
      uomType: "reference",
      rounding: "0.001",
    });
    setOpen(true);
  }

  function openEdit(row: UnitRow) {
    setEditing(row);
    form.reset({
      categoryId: row.categoryId,
      name: row.name,
      factor: String(row.factor),
      uomType: row.uomType,
      rounding: row.rounding,
    });
    setOpen(true);
  }

  function handleSubmit(values: UnitValues) {
    const payload = {
      categoryId: Number(values.categoryId),
      name: values.name,
      factor: Number(values.factor),
      uomType: values.uomType as Unit["uomType"],
      rounding: Number(values.rounding),
    };

    if (editing) {
      return getSwantaraService()
        .units.update(Number(editing.id), payload)
        .then(() => {
          toast.success(t("unitSaved"));
          setOpen(false);
          form.reset();
          void unitsQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update Unit", error);
          toast.error(t("toastFailed"));
        });
    } else {
      return getSwantaraService()
        .units.create(payload)
        .then(() => {
          toast.success(t("unitSaved"));
          setOpen(false);
          form.reset();
          void unitsQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create Unit", error);
          toast.error(t("toastFailed"));
        });
    }
  }

  function handleCategorySubmit(values: UnitCategoryValues) {
    return getSwantaraService()
      .uomCategories.create({ name: values.name })
      .then(() => {
        toast.success(t("categorySaved"));
        setCategoryOpen(false);
        categoryForm.reset();
        void categoriesQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to create Unit category", error);
        toast.error(t("toastFailed"));
      });
  }

  const columns: ColumnDef<UnitRow>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "categoryId",
      header: t("tableCategory"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {categories.find((category) => category.id === row.original.categoryId)?.name ?? "—"}
        </Badge>
      ),
    },
    {
      accessorKey: "factor",
      header: t("tableFactor"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.factor)}
        </span>
      ),
    },
    {
      accessorKey: "uomType",
      header: t("tableType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{String(row.original.uomType)}</span>
      ),
    },
    {
      accessorKey: "rounding",
      header: t("tableRounding"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{row.original.rounding}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("actionEdit")}
          deleteLabel={t("actionDelete")}
          confirmTitle={t("deleteUnitTitle")}
          confirmDescription={t("deleteUnitDescription")}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .units.delete(Number(row.original.id))
              .then(() => void unitsQuery.refetch())
              .catch((error) => {
                logger.error("Failed to delete Unit", error);
                toast.error(t("toastFailed"));
              })
          }
        />
      ),
    },
  ];

  const isLoading = categoriesQuery.isLoading || unitsQuery.isLoading;
  const error = categoriesQuery.isError
    ? categoriesQuery.error
    : unitsQuery.isError
      ? unitsQuery.error
      : null;

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle>{t("categoriesTitle")}</CardTitle>
            <CardDescription>{t("categoriesDescription")}</CardDescription>
          </div>
          <Button size="sm" variant="outline" onClick={() => setCategoryOpen(true)}>
            <Plus />
            <span>{t("addCategory")}</span>
          </Button>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {categoriesQuery.isLoading ? (
            <span className="text-sm text-muted-foreground">{t("loading")}</span>
          ) : categories.length === 0 ? (
            <span className="text-sm text-muted-foreground">{t("categoriesEmpty")}</span>
          ) : (
            categories.map((category) => (
              <Badge key={category.id} variant="secondary">
                {category.name}
              </Badge>
            ))
          )}
        </CardContent>
      </Card>

      <InteractiveEntityTable
        columns={columns}
        data={units}
        getRowId={(row) => row.id}
        searchKeys={["name"]}
        searchPlaceholder={t("searchUnitsPlaceholder")}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={t("unitsTitle")}
        emptyTitle={t("unitsEmpty")}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => {
                    void categoriesQuery.refetch();
                    void unitsQuery.refetch();
                  },
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={openCreate}>
            <Plus />
            <span>{t("addUnit")}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t("editUnit") : t("newUnit")}</DialogTitle>
            <DialogDescription>{t("unitDialogDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="categoryId" label={t("fieldCategory")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldCategory")}>
                      <SelectValue placeholder={t("fieldCategory")} />
                    </SelectTrigger>
                    <SelectContent>
                      {categories.map((category) => (
                        <SelectItem key={category.id} value={category.id}>
                          {category.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="uomType" label={t("fieldType")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldType")}>
                      <SelectValue placeholder={t("fieldType")} />
                    </SelectTrigger>
                    <SelectContent>
                      {uomTypes.map((type) => (
                        <SelectItem key={type} value={type}>
                          {String(type)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="name" label={t("fieldName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
              </FormField>
              <FormField name="factor" label={t("fieldFactor")}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="any" inputMode="decimal" />
                )}
              </FormField>
              <FormField name="rounding" label={t("fieldRounding")}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="any" inputMode="decimal" />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {t("actionCancel")}
              </Button>
              <SubmitButton>{t("saveUnit")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>

      <Dialog open={categoryOpen} onOpenChange={setCategoryOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("newCategory")}</DialogTitle>
          </DialogHeader>
          <Form form={categoryForm} onSubmit={handleCategorySubmit}>
            <FormField name="name" label={t("fieldName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
            </FormField>
            <DialogFooter>
              <Button variant="outline" onClick={() => setCategoryOpen(false)}>
                {t("actionCancel")}
              </Button>
              <SubmitButton>{t("saveCategory")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
