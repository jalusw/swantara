"use client";

import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { type ComponentProps, memo, useEffect, useMemo, useRef } from "react";
import { type DayButton, DayPicker, getDefaultClassNames, type Locale } from "react-day-picker";

import { cn } from "@/lib/utils";

const dayCellClassName =
  "relative flex size-11 items-center justify-center rounded-[var(--calendar-cell-radius)] p-0 tabular-nums outline-none after:absolute after:-inset-1 after:content-['']";

function Calendar({
  className,
  classNames,
  showOutsideDays = true,
  captionLayout = "label",
  locale,
  formatters,
  components,
  ...props
}: ComponentProps<typeof DayPicker>) {
  const defaultClassNames = useMemo(() => getDefaultClassNames(), []);

  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      captionLayout={captionLayout}
      locale={locale}
      formatters={{
        formatMonthDropdown: (date) => date.toLocaleString(locale?.code, { month: "short" }),
        ...formatters,
      }}
      className={cn(
        "w-fit [--calendar-cell-radius:var(--radius-md)]",
        "rtl:**:[.rdp-button_previous>svg]:rotate-180",
        "rtl:**:[.rdp-button_next>svg]:rotate-180",
        className,
      )}
      classNames={{
        root: cn("w-fit", defaultClassNames.root),
        months: cn("flex w-fit flex-col gap-6 md:flex-row md:gap-8", defaultClassNames.months),
        month: cn("flex w-fit flex-col gap-0.5", defaultClassNames.month),
        nav: cn(
          "absolute inset-x-0 top-0 flex w-full items-center justify-between px-1",
          defaultClassNames.nav,
        ),
        button_previous: cn(
          "relative inline-flex size-11 cursor-pointer items-center justify-center rounded-[var(--calendar-cell-radius)] text-muted-foreground transition-colors outline-none select-none after:absolute after:-inset-1 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:border-ring focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring aria-disabled:pointer-events-none aria-disabled:opacity-40",
          defaultClassNames.button_previous,
        ),
        button_next: cn(
          "relative inline-flex size-11 cursor-pointer items-center justify-center rounded-[var(--calendar-cell-radius)] text-muted-foreground transition-colors outline-none select-none after:absolute after:-inset-1 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:border-ring focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring aria-disabled:pointer-events-none aria-disabled:opacity-40",
          defaultClassNames.button_next,
        ),
        month_caption: cn(
          "flex h-11 w-full items-center justify-center",
          defaultClassNames.month_caption,
        ),
        caption_label: cn("text-sm text-foreground", defaultClassNames.caption_label),
        dropdowns: cn(
          "flex h-11 w-full items-center justify-center gap-1.5 text-sm",
          defaultClassNames.dropdowns,
        ),
        dropdown_root: cn(
          "relative rounded-[var(--calendar-cell-radius)]",
          defaultClassNames.dropdown_root,
        ),
        dropdown: cn(
          "absolute inset-0 cursor-pointer bg-background opacity-0",
          defaultClassNames.dropdown,
        ),
        month_grid: cn(
          "w-full border-separate border-spacing-0 border-collapse",
          defaultClassNames.month_grid,
        ),
        weekdays: cn("flex h-9 items-center", defaultClassNames.weekdays),
        weekday: cn("flex-1 text-center text-xs text-muted-foreground", defaultClassNames.weekday),
        week: cn("flex w-full items-center", defaultClassNames.week),
        week_number_header: cn("w-6 select-none", defaultClassNames.week_number_header),
        week_number: cn(
          "w-6 text-center text-xs text-muted-foreground select-none",
          defaultClassNames.week_number,
        ),
        day: cn(dayCellClassName, defaultClassNames.day),
        range_start: cn(
          "relative isolate rounded-l-[var(--calendar-cell-radius)] bg-muted",
          defaultClassNames.range_start,
        ),
        range_middle: cn("bg-muted", defaultClassNames.range_middle),
        range_end: cn(
          "relative isolate rounded-r-[var(--calendar-cell-radius)] bg-muted",
          defaultClassNames.range_end,
        ),
        outside: cn("text-muted-foreground opacity-40", defaultClassNames.outside),
        disabled: cn("text-muted-foreground opacity-50", defaultClassNames.disabled),
        hidden: cn("invisible", defaultClassNames.hidden),
        ...classNames,
      }}
      components={{
        Root: ({ className, rootRef, ...props }) => {
          return <div data-slot="calendar" ref={rootRef} className={cn(className)} {...props} />;
        },
        Chevron: ({ className, orientation, ...props }) => {
          if (orientation === "left") {
            return (
              <ChevronLeftIcon
                data-slot="calendar-chevron-left"
                className={cn("size-4", className)}
                {...props}
              />
            );
          }
          if (orientation === "right") {
            return (
              <ChevronRightIcon
                data-slot="calendar-chevron-right"
                className={cn("size-4", className)}
                {...props}
              />
            );
          }
          return (
            <ChevronDownIcon
              data-slot="calendar-chevron-down"
              className={cn("size-4", className)}
              {...props}
            />
          );
        },
        WeekNumber: ({ children, ...props }) => {
          return (
            <td {...props}>
              <div className="flex size-9 items-center justify-center text-center">{children}</div>
            </td>
          );
        },
        ...components,
      }}
      {...props}
    />
  );
}

const CalendarDayButton = memo(function CalendarDayButton({
  className,
  day,
  modifiers,
  locale,
  ...props
}: ComponentProps<typeof DayButton> & {
  locale?: Partial<Locale>;
}) {
  const defaultClassNames = useMemo(() => getDefaultClassNames(), []);

  const ref = useRef<HTMLButtonElement>(null);
  const isFocused = modifiers.focused;
  useEffect(() => {
    if (isFocused) {
      ref.current?.focus({ preventScroll: true });
    }
  }, [isFocused]);

  const isToday = modifiers.today && !modifiers.selected;

  return (
    <button
      ref={ref}
      type="button"
      data-day={day.date.toLocaleDateString(locale?.code)}
      data-selected-single={
        modifiers.selected &&
        !modifiers.range_start &&
        !modifiers.range_end &&
        !modifiers.range_middle
      }
      data-range-start={modifiers.range_start}
      data-range-end={modifiers.range_end}
      data-range-middle={modifiers.range_middle}
      data-today={isToday}
      className={cn(
        "relative inline-flex size-11 cursor-pointer items-center justify-center rounded-[var(--calendar-cell-radius)] p-0 text-sm font-normal text-foreground tabular-nums transition-colors outline-none select-none after:absolute after:-inset-1 after:content-['']",
        "hover:bg-muted hover:text-foreground",
        "focus-visible:border-ring focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring",
        "aria-disabled:pointer-events-none aria-disabled:opacity-40",
        "data-[selected-single=true]:bg-primary data-[selected-single=true]: data-[selected-single=true]:text-primary-foreground",
        "data-[range-start=true]:data-[range-end=true]:rounded-[var(--calendar-cell-radius)]",
        "data-[range-start=true]:rounded-l-[var(--radius-lg)] data-[range-start=true]:bg-primary data-[range-start=true]: data-[range-start=true]:text-primary-foreground",
        "data-[range-end=true]:rounded-r-[var(--radius-lg)] data-[range-end=true]:bg-primary data-[range-end=true]: data-[range-end=true]:text-primary-foreground",
        "data-[range-middle=true]:rounded-none data-[range-middle=true]:bg-muted data-[range-middle=true]:text-foreground hover:data-[range-middle=true]:bg-muted",
        "data-[today=true]:bg-muted data-[today=true]:text-foreground data-[today=true]:hover:bg-muted hover:data-[today=true]:bg-muted",
        defaultClassNames.day,
        className,
      )}
      {...props}
    />
  );
});

export { Calendar, CalendarDayButton };
