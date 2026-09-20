"use client";

import { CalendarDaysIcon } from "lucide-react";
import { type ComponentProps, useState } from "react";
import type { DateRange, OnSelectHandler } from "react-day-picker";

import { Button } from "./button";
import { Calendar } from "./calendar";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";

function formatDate(date: Date, locale?: string) {
  return new Intl.DateTimeFormat(locale, { dateStyle: "medium" }).format(date);
}

type DatePickerProps = Omit<ComponentProps<typeof Calendar>, "mode" | "selected" | "onSelect"> & {
  selected?: Date | undefined;
  onSelect?: OnSelectHandler<Date | undefined> | undefined;
  placeholder?: string;
  className?: string;
};

function DatePicker({
  className,
  locale,
  selected,
  onSelect,
  placeholder = "Pick a date",
  ...props
}: DatePickerProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover data-slot="date-picker" open={open} onOpenChange={setOpen}>
      <PopoverTrigger render={<Button variant="outline" className={className} />}>
        <CalendarDaysIcon className="text-muted-foreground" />
        {selected ? formatDate(selected, locale?.code) : placeholder}
      </PopoverTrigger>
      <PopoverContent align="start" className="w-auto p-0" aria-label={"Choose a date"}>
        <Calendar
          mode="single"
          locale={locale}
          selected={selected}
          onSelect={(day, triggerDate, modifiers, e) => {
            onSelect?.(day, triggerDate, modifiers, e);
            setOpen(false);
          }}
          {...props}
        />
      </PopoverContent>
    </Popover>
  );
}

type DateRangePickerProps = Omit<
  ComponentProps<typeof Calendar>,
  "mode" | "selected" | "onSelect"
> & {
  selected?: DateRange | undefined;
  onSelect?: OnSelectHandler<DateRange | undefined> | undefined;
  placeholder?: string;
  className?: string;
};

function DateRangePicker({
  className,
  locale,
  selected,
  onSelect,
  placeholder = "Pick a date range",
  ...props
}: DateRangePickerProps) {
  const [open, setOpen] = useState(false);

  const from = selected?.from;
  const to = selected?.to;
  const fromLabel = from ? formatDate(from, locale?.code) : undefined;
  const toLabel = to ? formatDate(to, locale?.code) : undefined;

  const rangeLabel = to ? `${fromLabel} – ${toLabel}` : from ? `${fromLabel} – …` : placeholder;

  return (
    <Popover data-slot="date-range-picker" open={open} onOpenChange={setOpen}>
      <PopoverTrigger render={<Button variant="outline" className={className} />}>
        <CalendarDaysIcon className="text-muted-foreground" />
        {rangeLabel}
      </PopoverTrigger>
      <PopoverContent align="start" className="w-auto p-0" aria-label={"Choose a date range"}>
        <Calendar mode="range" locale={locale} selected={selected} onSelect={onSelect} {...props} />
      </PopoverContent>
    </Popover>
  );
}

export { DatePicker, DateRangePicker };
export type { DateRange };
