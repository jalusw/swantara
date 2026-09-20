"use client";

import { Plus, X } from "lucide-react";
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
            placeholder={"Attribute"}
            aria-label={`${"Attribute"} ${attribute.name || ""}`}
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
            placeholder={"Values (comma-separated)"}
            aria-label={`${"Values (comma-separated)"} ${attribute.name || ""}`}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => onChange(value.filter((entry) => entry.id !== attribute.id))}
            aria-label={"Remove"}
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
          <span>{"Add attribute"}</span>
        </Button>
      </div>
    </div>
  );
}
