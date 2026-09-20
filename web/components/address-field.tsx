"use client";

import { useId } from "react";
import { cn } from "@/lib/utils";
import { Field } from "./field";
import { Input } from "./input";
import { Label } from "./label";

export type AddressValue = {
  address: string;
  city: string;
  postalCode: string;
};

export type AddressFieldLabels = {
  address?: string;
  city?: string;
  postalCode?: string;
};

export type AddressFieldProps = {
  value: AddressValue;
  onChange: (value: AddressValue) => void;
  disabled?: boolean;
  className?: string;
  labels?: AddressFieldLabels;
};

function patch(value: AddressValue, field: keyof AddressValue, next: string) {
  return { ...value, [field]: next };
}

export function AddressField({
  value,
  onChange,
  disabled = false,
  className,
  labels,
}: AddressFieldProps) {
  const allLabels = {
    address: "Street address",
    city: "City",
    postalCode: "Postal code",
    ...labels,
  };
  const rootId = useId();

  return (
    <div data-slot="address-field" className={cn("grid gap-3", className)}>
      <Field>
        <Label htmlFor={`${rootId}-address`}>{allLabels.address}</Label>
        <Input
          id={`${rootId}-address`}
          value={value.address}
          onChange={(event) => onChange(patch(value, "address", event.target.value))}
          disabled={disabled}
        />
      </Field>

      <div className="grid gap-3 sm:grid-cols-2">
        <Field>
          <Label htmlFor={`${rootId}-city`}>{allLabels.city}</Label>
          <Input
            id={`${rootId}-city`}
            value={value.city}
            onChange={(event) => onChange(patch(value, "city", event.target.value))}
            disabled={disabled}
          />
        </Field>
        <Field>
          <Label htmlFor={`${rootId}-postalCode`}>{allLabels.postalCode}</Label>
          <Input
            id={`${rootId}-postalCode`}
            value={value.postalCode}
            onChange={(event) => onChange(patch(value, "postalCode", event.target.value))}
            disabled={disabled}
          />
        </Field>
      </div>
    </div>
  );
}
