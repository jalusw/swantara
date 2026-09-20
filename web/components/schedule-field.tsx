"use client";

import { cn } from "@/lib/utils";
import { Input } from "./input";
import { Label } from "./label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";
import { ToggleGroup, ToggleGroupItem } from "./toggle-group";

export type ScheduleType = "once" | "daily" | "weekly" | "monthly";

export type Schedule = {
  type: ScheduleType;
  interval: number;
  weekdays: number[];
};

export const WEEKDAY_SHORT: Record<number, string> = {
  1: "Mon",
  2: "Tue",
  3: "Wed",
  4: "Thu",
  5: "Fri",
  6: "Sat",
  7: "Sun",
};

export type ScheduleFieldProps = {
  value: Schedule;
  onChange: (value: Schedule) => void;
  className?: string;
};

const typeOptions: ScheduleType[] = ["once", "daily", "weekly", "monthly"];

export function ScheduleField({ value, onChange, className }: ScheduleFieldProps) {
  const weekdays = Object.keys(WEEKDAY_SHORT).map((key) => Number(key));

  return (
    <div data-slot="schedule-field" className={cn("flex flex-col gap-4", className)}>
      <div className="flex flex-wrap items-end gap-3">
        <div className="flex min-w-44 flex-1 flex-col gap-1.5">
          <Label htmlFor="schedule-type">{"Repeats"}</Label>
          <Select
            value={value.type}
            onValueChange={(type) => onChange({ ...value, type: type as ScheduleType })}
          >
            <SelectTrigger id="schedule-type" aria-label={"Repeats"} className="w-full">
              <SelectValue placeholder="Select frequency" />
            </SelectTrigger>
            <SelectContent>
              {typeOptions.map((type) => (
                <SelectItem key={type} value={type}>
                  {String(type)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {value.type !== "once" ? (
          <div className="flex max-w-28 flex-1 flex-col gap-1.5">
            <Label htmlFor="schedule-interval">{"Every"}</Label>
            <Input
              id="schedule-interval"
              type="number"
              min={1}
              value={value.interval}
              onChange={(event) =>
                onChange({
                  ...value,
                  interval: Math.max(1, Number(event.target.value) || 1),
                })
              }
            />
          </div>
        ) : null}
      </div>

      {value.type === "weekly" ? (
        <div className="flex flex-col gap-1.5">
          <Label>{"On days"}</Label>
          <ToggleGroup
            multiple
            aria-label={"On days"}
            value={value.weekdays.map(String)}
            onValueChange={(days) =>
              onChange({
                ...value,
                weekdays: days.map(Number).sort((a, b) => a - b),
              })
            }
          >
            {weekdays.map((day) => (
              <ToggleGroupItem key={day} value={String(day)}>
                {WEEKDAY_SHORT[day]}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
        </div>
      ) : null}
    </div>
  );
}
