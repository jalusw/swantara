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
  return z.object({
    categoryId: z.string().min(1, "Select a category."),
    name: z.string().min(1, "Enter a name."),
    factor: z
      .string()
      .min(1, "Enter a factor.")
      .refine((value) => Number(value) > 0, "Factor must be greater than zero."),
    uomType: z.enum(uomTypes),
    rounding: z.string(),
  });
}
type UnitValues = z.infer<ReturnType<typeof useUnitSchema>>;

function useUnitCategorySchema() {
  return z.object({
    name: z.string().min(1, "Enter a category name."),
  });
}
type UnitCategoryValues = z.infer<ReturnType<typeof useUnitCategorySchema>>;

export function UnitsSection(_props: { orgId: string }) {
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
      void getSwantaraService()
        .units.update(Number(editing.id), payload)
        .then(() => {
          toast.success("Unit of measure saved.");
          setOpen(false);
          form.reset();
          void unitsQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update Unit", error);
          toast.error("Could not disable the organization.");
        });
    } else {
      void getSwantaraService()
        .units.create(payload)
        .then(() => {
          toast.success("Unit of measure saved.");
          setOpen(false);
          form.reset();
          void unitsQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create Unit", error);
          toast.error("Could not disable the organization.");
        });
    }
  }

  function handleCategorySubmit(values: UnitCategoryValues) {
    void getSwantaraService()
      .uomCategories.create({ name: values.name })
      .then(() => {
        toast.success("Category saved.");
        setCategoryOpen(false);
        categoryForm.reset();
        void categoriesQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to create Unit category", error);
        toast.error("Could not disable the organization.");
      });
  }

  const columns: ColumnDef<UnitRow>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "categoryId",
      header: "Category",
      cell: ({ row }) => (
        <Badge variant="secondary">
          {categories.find((category) => category.id === row.original.categoryId)?.name ?? "—"}
        </Badge>
      ),
    },
    {
      accessorKey: "factor",
      header: "Factor",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.factor)}
        </span>
      ),
    },
    {
      accessorKey: "uomType",
      header: "Type",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{String(row.original.uomType)}</span>
      ),
    },
    {
      accessorKey: "rounding",
      header: "Rounding",
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
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this UoM?"}
          confirmDescription={"The unit of measure will be removed from the catalog."}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .units.delete(Number(row.original.id))
              .then(() => void unitsQuery.refetch())
              .catch((error) => {
                logger.error("Failed to delete Unit", error);
                toast.error("Could not disable the organization.");
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
            <CardTitle>{"Categories"}</CardTitle>
            <CardDescription>{"Groups of compatible units."}</CardDescription>
          </div>
          <Button size="sm" variant="outline" onClick={() => setCategoryOpen(true)}>
            <Plus />
            <span>{"Add category"}</span>
          </Button>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {categoriesQuery.isLoading ? (
            <span className="text-sm text-muted-foreground">{"Loading..."}</span>
          ) : categories.length === 0 ? (
            <span className="text-sm text-muted-foreground">{"No categories yet"}</span>
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
        searchPlaceholder={"Search units…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"Units of measure"}
        emptyTitle={"No units of measure"}
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
            <span>{"Add UoM"}</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? "Edit unit of measure" : "New unit of measure"}</DialogTitle>
            <DialogDescription>
              {"Units of measure and their categories used across products and stock."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="categoryId" label={"Category"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Category"}>
                      <SelectValue placeholder={"Category"} />
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
              <FormField name="uomType" label={"Type"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Type"}>
                      <SelectValue placeholder={"Type"} />
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
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="factor" label={"Factor"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="any" inputMode="decimal" />
                )}
              </FormField>
              <FormField name="rounding" label={"Rounding"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="any" inputMode="decimal" />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save unit"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>

      <Dialog open={categoryOpen} onOpenChange={setCategoryOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{"New category"}</DialogTitle>
          </DialogHeader>
          <Form form={categoryForm} onSubmit={handleCategorySubmit}>
            <FormField name="name" label={"Name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
            </FormField>
            <DialogFooter>
              <Button variant="outline" onClick={() => setCategoryOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save category"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
