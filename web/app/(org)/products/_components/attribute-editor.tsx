"use client";

import { Plus, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { Input } from "@/components/input";

export type AttributeOption = {
  name: string;
  values: string[];
};

export type AttributeRow = AttributeOption & { id: string };

function createRow(): AttributeRow {
  return {
    id: crypto.randomUUID(),
    name: "",
    values: [],
  };
}

export function AttributeEditor({
  value,
  onChange,
}: {
  value: AttributeRow[];
  onChange: (attributes: AttributeRow[]) => void;
}) {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Products");
  const tCommon = useTranslations("Common");
  function update(id: string, patch: Partial<AttributeOption>) {
    onChange(
      value.map((attribute) => (attribute.id === id ? { ...attribute, ...patch } : attribute)),
    );
  }

  return (
    <div className="flex flex-col gap-2">
      {value.map((attribute) => (
        <div key={attribute.id} className="flex items-center gap-2">
          <Input
            value={attribute.name}
            onChange={(event) => update(attribute.id, { name: event.target.value })}
            placeholder={t("attribute")}
            aria-label={`${t("attribute")} ${attribute.name || ""}`}
            className="w-40"
          />
          <Input
            value={attribute.values.join(", ")}
            onChange={(event) =>
              update(attribute.id, {
                values: event.target.value
                  .split(",")
                  .map((entry) => entry.trim())
                  .filter(Boolean),
              })
            }
            placeholder={t("attributeValuesPlaceholder")}
            aria-label={`${t("attributeValuesPlaceholder")} ${attribute.name || ""}`}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => onChange(value.filter((entry) => entry.id !== attribute.id))}
            aria-label={tCommon("delete")}
          >
            <X />
          </Button>
        </div>
      ))}
      <div>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => onChange([...value, createRow()])}
        >
          <Plus />
          <span>{t("addAttribute")}</span>
        </Button>
      </div>
    </div>
  );
}
