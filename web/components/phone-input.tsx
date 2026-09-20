"use client";

import { Combobox as ComboboxPrimitive } from "@base-ui/react";
import getCountryFlag from "country-flag-icons/unicode";
import { ChevronDownIcon } from "lucide-react";
import { type ComponentProps, useState } from "react";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxItem,
  ComboboxList,
  useComboboxAnchor,
} from "./combobox";
import { Input } from "./input";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "./input-group";

export type PhoneInputProps = ComponentProps<"input"> & {
  defaultCountryCode?: string;
};

type Country = {
  iso: string;
  name: string;
  dial: string;
};

export const COUNTRIES: Country[] = [
  { iso: "id", name: "Indonesia", dial: "62" },
  { iso: "us", name: "United States", dial: "1" },
  { iso: "ca", name: "Canada", dial: "1" },
  { iso: "sg", name: "Singapore", dial: "65" },
  { iso: "my", name: "Malaysia", dial: "60" },
  { iso: "th", name: "Thailand", dial: "66" },
  { iso: "vn", name: "Vietnam", dial: "84" },
  { iso: "ph", name: "Philippines", dial: "63" },
  { iso: "cn", name: "China", dial: "86" },
  { iso: "jp", name: "Japan", dial: "81" },
  { iso: "kr", name: "South Korea", dial: "82" },
  { iso: "in", name: "India", dial: "91" },
  { iso: "au", name: "Australia", dial: "61" },
  { iso: "nz", name: "New Zealand", dial: "64" },
  { iso: "gb", name: "United Kingdom", dial: "44" },
  { iso: "de", name: "Germany", dial: "49" },
  { iso: "fr", name: "France", dial: "33" },
  { iso: "it", name: "Italy", dial: "39" },
  { iso: "es", name: "Spain", dial: "34" },
  { iso: "nl", name: "Netherlands", dial: "31" },
  { iso: "br", name: "Brazil", dial: "55" },
  { iso: "mx", name: "Mexico", dial: "52" },
  { iso: "tr", name: "Turkey", dial: "90" },
  { iso: "sa", name: "Saudi Arabia", dial: "966" },
  { iso: "ae", name: "United Arab Emirates", dial: "971" },
];

const DEFAULT_COUNTRY_ISO = "id";

export function getDialCodePrefix(value: string | undefined): string {
  const trimmed = value?.trim() ?? "";
  return (
    COUNTRIES.find(
      (country) => trimmed === `+${country.dial}` || trimmed.startsWith(`+${country.dial} `),
    )?.dial ?? ""
  );
}

function findCountry(iso: string | undefined): Country {
  return (
    COUNTRIES.find((c) => c.iso === iso) ??
    COUNTRIES.find((c) => c.iso === DEFAULT_COUNTRY_ISO) ??
    COUNTRIES[0]!
  );
}

function PhoneInput({ defaultCountryCode, ...props }: PhoneInputProps) {
  const initialValue = props.defaultValue ?? props.value ?? "";
  const detectedDial = typeof initialValue === "string" ? getDialCodePrefix(initialValue) : "";
  const initialCountryIso =
    (detectedDial && COUNTRIES.find((c) => c.dial === detectedDial)?.iso) ??
    defaultCountryCode ??
    DEFAULT_COUNTRY_ISO;

  const [countryCode, setCountryCode] = useState(initialCountryIso);
  const selectedCountry = findCountry(countryCode);
  const anchor = useComboboxAnchor();

  return (
    <Combobox
      data-slot="phone-input"
      value={countryCode}
      onValueChange={(val) => {
        if (val) {
          setCountryCode(val);
        }
      }}
      itemToStringLabel={(val) => {
        const country = COUNTRIES.find((c) => c.iso === val);
        return country ? `${country.name} ${country.dial}` : String(val ?? "");
      }}
    >
      <InputGroup className="w-full" ref={anchor}>
        <InputGroupAddon align="inline-start" className="pl-1">
          <ComboboxPrimitive.Trigger
            render={<InputGroupButton variant="ghost" size="sm" className="gap-1.5 px-2" />}
            aria-label={`${selectedCountry.name}, dial code +${selectedCountry.dial}`}
          >
            <span className="text-base leading-none">{getCountryFlag(selectedCountry.iso)}</span>
            <span className="font-mono text-xs">+{selectedCountry.dial}</span>
            <ChevronDownIcon className="size-3.5 text-muted-foreground" />
          </ComboboxPrimitive.Trigger>
        </InputGroupAddon>
        <InputGroupInput type="tel" placeholder="+123456789" {...props} />
      </InputGroup>
      <ComboboxContent anchor={anchor}>
        <ComboboxPrimitive.Input
          render={
            <Input
              className="min-h-11 border-border/60 bg-background"
              placeholder={"Search country…"}
            />
          }
          aria-label={"Search country…"}
        />
        <ComboboxList>
          {COUNTRIES.map((country) => (
            <ComboboxItem key={country.iso} value={country.iso}>
              <span className="text-base leading-none">{getCountryFlag(country.iso)}</span>
              <span className="flex-1">{country.name}</span>
              <span className="font-mono text-xs text-muted-foreground">+{country.dial}</span>
            </ComboboxItem>
          ))}
          <ComboboxEmpty>{"No countries found"}</ComboboxEmpty>
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  );
}

export { PhoneInput };
