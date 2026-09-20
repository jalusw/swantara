import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ScheduleField } from "@/components/schedule-field";
import { renderWithProviders } from "@/lib/tests";

const baseSchedule = { type: "once" as const, interval: 1, weekdays: [] };

describe("ScheduleField", () => {
  it("renders with data-slot", () => {
    renderWithProviders(<ScheduleField value={baseSchedule} onChange={vi.fn()} />);

    expect(document.querySelector("[data-slot='schedule-field']")).toBeInTheDocument();
  });

  it("renders the schedule type select", () => {
    renderWithProviders(<ScheduleField value={baseSchedule} onChange={vi.fn()} />);

    expect(screen.getByLabelText("Repeats")).toBeInTheDocument();
  });

  it("hides interval input when type is once", () => {
    renderWithProviders(<ScheduleField value={baseSchedule} onChange={vi.fn()} />);

    expect(screen.queryByLabelText("Every")).not.toBeInTheDocument();
  });
});
