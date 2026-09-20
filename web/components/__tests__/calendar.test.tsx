import { render, screen } from "@testing-library/react";
import { CalendarDay } from "react-day-picker";
import { describe, expect, it } from "vitest";
import { Calendar, CalendarDayButton } from "@/components/calendar";

describe("Calendar", () => {
  it("renders a month grid with navigation chevrons", () => {
    render(<Calendar mode="single" selected={new Date(2026, 0, 15)} />);

    expect(screen.getByRole("grid")).toBeInTheDocument();
    expect(screen.queryByTestId("calendar-chevron-left")).not.toBeInTheDocument();
  });

  it("renders with month and year dropdowns", () => {
    render(<Calendar mode="single" captionLayout="dropdown" />);

    expect(screen.getByRole("grid")).toBeInTheDocument();
  });
});

describe("CalendarDayButton", () => {
  const day = new CalendarDay(new Date(2026, 0, 15), new Date(2026, 0, 1));

  it("renders a plain day", () => {
    render(<CalendarDayButton day={day} modifiers={{}} locale={{ code: "en" }} />);

    expect(screen.getByRole("button")).toBeInTheDocument();
    expect(screen.getByRole("button")).toHaveAttribute("data-day");
  });

  it("marks a focused today button", () => {
    render(
      <CalendarDayButton
        day={day}
        modifiers={{ focused: true, today: true }}
        locale={{ code: "en" }}
      />,
    );

    expect(screen.getByRole("button")).toHaveAttribute("data-today", "true");
  });

  it("marks a selected single day", () => {
    render(<CalendarDayButton day={day} modifiers={{ selected: true }} locale={{ code: "en" }} />);

    expect(screen.getByRole("button")).toHaveAttribute("data-selected-single", "true");
  });

  it("marks a range start day as not selected-single", () => {
    render(
      <CalendarDayButton
        day={day}
        modifiers={{ selected: true, range_start: true }}
        locale={{ code: "en" }}
      />,
    );

    expect(screen.getByRole("button")).toHaveAttribute("data-selected-single", "false");
    expect(screen.getByRole("button")).toHaveAttribute("data-range-start", "true");
  });

  it("marks a range end day", () => {
    render(
      <CalendarDayButton
        day={day}
        modifiers={{ selected: true, range_end: true }}
        locale={{ code: "en" }}
      />,
    );
    expect(screen.getByRole("button")).toHaveAttribute("data-range-end", "true");
  });

  it("marks a range middle day", () => {
    render(
      <CalendarDayButton
        day={day}
        modifiers={{ selected: true, range_middle: true }}
        locale={{ code: "en" }}
      />,
    );
    expect(screen.getByRole("button")).toHaveAttribute("data-range-middle", "true");
  });
});
