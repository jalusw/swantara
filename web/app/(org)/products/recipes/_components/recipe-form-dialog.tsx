"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, X } from "lucide-react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Switch } from "@/components/switch";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Unit } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { bomTypes, type StubItem } from "../../_components/products-data";

const lineSchemaShape = () =>
  z.object({
    id: z.string(),
    componentId: z.string().min(1, "Select a component."),
    qty: z
      .string()
      .min(1, "Enter a quantity.")
      .refine((value) => Number(value) > 0, "Quantity must be greater than zero."),
    unitId: z.string(),
    scrapPct: z.string().refine((value) => Number(value) >= 0, "Scrap must be zero or more."),
  });

const buildSchema = () =>
  z.object({
    itemId: z.string().min(1, "Select a item."),
    code: z.string(),
    type: z.enum(bomTypes),
    qty: z
      .string()
      .min(1, "Enter a quantity.")
      .refine((value) => Number(value) > 0, "Quantity must be greater than zero."),
    unitId: z.string(),
    version: z
      .string()
      .min(1, "Enter a version number.")
      .refine((value) => Number(value) >= 1, "Enter a version number."),
    active: z.boolean(),
    lines: z.array(lineSchemaShape()).min(1, "Add at least one component."),
  });

type Values = z.infer<ReturnType<typeof buildSchema>>;
type LineValues = z.infer<ReturnType<typeof lineSchemaShape>>;

function bomLineTotalQty(line: { qty: number; scrapPct: number }): number {
  return line.qty * (1 + line.scrapPct / 100);
}

export function BomFormDialog({
  open,
  onOpenChange,
  orgId,
  templates,
  initial,
  initialLines,
  nextVersion = 1,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  templates: StubItem[];
  initial?: {
    id: string;
    itemId: string;
    code: string | null;
    type: Values["type"];
    qty: number;
    unitId: string | null;
    version: number;
    active: boolean;
  };
  initialLines?: {
    id: string;
    componentId: string;
    qty: number;
    unitId: string | null;
    scrapPct: number;
  }[];
  nextVersion?: number;
  onSave: () => void;
}) {
  const unitsQuery = useOrgListQuery<{ units: Unit[] }, Record<string, never>>("units", () =>
    getSwantaraService().units.list(),
  );
  const uomOptions = (unitsQuery.data?.units ?? []).map((unit) => ({
    id: String(unit.id),
    name: unit.name,
  }));

  const form = useForm<Values>({
    resolver: zodResolver(buildSchema()),
    defaultValues: initial ? toValues(initial, initialLines ?? []) : emptyValues(nextVersion),
  });
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });
  const lineValues = form.watch("lines");

  function handleSubmit(values: Values) {
    const numId = initial ? Number(initial.id) || 0 : 0;

    if (initial) {
      void getSwantaraService()
        .recipes.update(Number(orgId) || 0, numId, {
          organizationId: Number(orgId) || 0,
          itemId: Number(values.itemId) || 0,
          code: values.code || null,
          type: values.type,
          qty: Number(values.qty) || 0,
          unitId: values.unitId === "" ? null : Number(values.unitId) || null,
          version: Number(values.version) || 1,
          active: values.active,
        })
        .then(() => {
          toast.success("Bill of materials saved.");
          onSave();
        })
        .catch(() => toast.error("Something went wrong. Please try again."));
    } else {
      void getSwantaraService()
        .recipes.create(Number(orgId) || 0, {
          organizationId: Number(orgId) || 0,
          itemId: Number(values.itemId) || 0,
          code: values.code || null,
          type: values.type,
          qty: Number(values.qty) || 0,
          unitId: values.unitId === "" ? null : Number(values.unitId) || null,
          version: Number(values.version) || 1,
          lines: values.lines.map((line) => ({
            componentId: Number(line.componentId) || 0,
            qty: Number(line.qty) || 0,
            unitId: line.unitId === "" ? null : Number(line.unitId) || null,
            scrapPct: Number(line.scrapPct || 0),
          })),
        })
        .then(() => {
          toast.success("Bill of materials saved.");
          onSave();
        })
        .catch(() => toast.error("Something went wrong. Please try again."));
    }
  }

  const totalWithScrap = (lineValues ?? []).reduce(
    (sum, line) =>
      sum +
      (Number(line?.qty) > 0
        ? bomLineTotalQty({
            qty: Number(line.qty),
            scrapPct: Number(line.scrapPct || 0),
          })
        : 0),
    0,
  );

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={initial ? "Edit bill of materials" : "New bill of materials"}
      description={
        "Recipes that define how manufactured, kit, and subcontracted products are assembled."
      }
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="itemId" label={"Item"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Item"}>
                <SelectValue placeholder={"Item"} />
              </SelectTrigger>
              <SelectContent>
                {templates.map((template) => (
                  <SelectItem key={template.id} value={template.id}>
                    {template.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="code" label={"Reference"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"BOM/..."} />}
        </FormField>
        <FormField name="type" label={"Type"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Type"}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {bomTypes.map((bomType) => (
                  <SelectItem key={bomType} value={bomType}>
                    {humanizeKey(String(bomType))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="version" label={"Version"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="1" step="1" />}
        </FormField>
        <FormField name="qty" label={"Quantity"}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0" step="any" inputMode="decimal" />
          )}
        </FormField>
        <FormField name="unitId" label={"Unit of measure"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Unit of measure"}>
                <SelectValue placeholder={"Unit of measure"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {uomOptions.map((unit) => (
                  <SelectItem key={unit.id} value={unit.id}>
                    {unit.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="active" label={"Active"}>
          {({ field }) => (
            <Switch
              checked={field.value}
              onCheckedChange={(checked) => field.onChange(Boolean(checked))}
              aria-label={"Active"}
            />
          )}
        </FormField>
      </div>

      <div className="flex flex-col gap-2">
        <p className="text-sm">{"Components"}</p>
        <p className="text-xs text-muted-foreground">
          {"Components consumed per produced quantity, including scrap allowance."}
        </p>
        {fields.map((field, index) => (
          <div key={field.id} className="flex items-start gap-2">
            <FormField name={`lines.${index}.componentId`}>
              {({ field: lineField, id }) => (
                <Select value={lineField.value} onValueChange={lineField.onChange}>
                  <SelectTrigger id={id} aria-label={"Component"}>
                    <SelectValue placeholder={"Component"} />
                  </SelectTrigger>
                  <SelectContent>
                    {templates.map((template) => (
                      <SelectItem key={template.id} value={template.id}>
                        {template.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name={`lines.${index}.qty`}>
              {({ field: lineField, id }) => (
                <Input
                  {...lineField}
                  id={id}
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  placeholder={"Quantity"}
                  aria-label={"Quantity"}
                  className="w-24"
                />
              )}
            </FormField>
            <FormField name={`lines.${index}.scrapPct`}>
              {({ field: lineField, id }) => (
                <Input
                  {...lineField}
                  id={id}
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  placeholder={"Scrap %"}
                  aria-label={"Scrap %"}
                  className="w-20"
                />
              )}
            </FormField>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => remove(index)}
              aria-label={"Remove line"}
            >
              <X />
            </Button>
          </div>
        ))}
        <div className="flex items-center justify-between">
          <Button type="button" size="sm" variant="outline" onClick={() => append(emptyLine())}>
            <Plus />
            <span>{"Add component"}</span>
          </Button>
          <span className="text-sm text-muted-foreground">
            {`Total incl. scrap: ${formatNumber(totalWithScrap)}`}
          </span>
        </div>
      </div>
    </EntityFormDialog>
  );
}

function emptyLine(): LineValues {
  return {
    id: `line-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    componentId: "",
    qty: "",
    unitId: "",
    scrapPct: "0",
  };
}

function emptyValues(nextVersion: number): Values {
  return {
    itemId: "",
    code: "",
    type: "manufacture",
    qty: "1",
    unitId: "",
    version: String(nextVersion),
    active: true,
    lines: [emptyLine()],
  };
}

function toValues(
  recipe: {
    itemId: string;
    code: string | null;
    type: Values["type"];
    qty: number;
    unitId: string | null;
    version: number;
    active: boolean;
  },
  lines: {
    id: string;
    componentId: string;
    qty: number;
    unitId: string | null;
    scrapPct: number;
  }[],
): Values {
  return {
    itemId: recipe.itemId,
    code: recipe.code ?? "",
    type: recipe.type,
    qty: String(recipe.qty),
    unitId: recipe.unitId ?? "",
    version: String(recipe.version),
    active: recipe.active,
    lines: lines.map((line) => ({
      id: line.id,
      componentId: line.componentId,
      qty: String(line.qty),
      unitId: line.unitId ?? "",
      scrapPct: String(line.scrapPct),
    })),
  };
}
