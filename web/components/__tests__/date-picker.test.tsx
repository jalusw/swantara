import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DatePicker, DateRangePicker } from "@/components/date-picker";
import { renderWithProviders } from "@/lib/tests";

describe("DatePicker", () => {
  it("opens the calendar popup and selects a single date", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(
      <DatePicker selected={undefined} onSelect={onSelect} defaultMonth={new Date(2026, 7)} />,
    );

    await user.click(screen.getByRole("button", { name: "Pick a date" }));
    expect(await screen.findByRole("grid")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /August 15th/ }));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it("renders the selected date in the trigger", () => {
    const date = new Date(2026, 7, 15);
    renderWithProviders(<DatePicker selected={date} />);

    expect(screen.getByRole("button", { name: /Aug 15, 2026/ })).toBeInTheDocument();
  });
});

describe("DateRangePicker", () => {
  it("opens the calendar popup and selects a range", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(
      <DateRangePicker selected={undefined} onSelect={onSelect} defaultMonth={new Date(2026, 7)} />,
    );

    await user.click(screen.getByRole("button", { name: "Pick a date range" }));
    expect(await screen.findByRole("grid")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /August 15th/ }));
    await user.click(screen.getByRole("button", { name: /August 20th/ }));
    expect(onSelect).toHaveBeenCalledTimes(2);
    expect(onSelect).toHaveBeenLastCalledWith(
      expect.objectContaining({
        from: expect.any(Date),
        to: expect.any(Date),
      }),
      expect.any(Date),
      expect.any(Object),
      expect.anything(),
    );
  });

  it("renders the selected range in the trigger", () => {
    renderWithProviders(
      <DateRangePicker selected={{ from: new Date(2026, 7, 15), to: new Date(2026, 7, 20) }} />,
    );

    expect(screen.getByRole("button", { name: /Aug 15, 2026.*Aug 20, 2026/ })).toBeInTheDocument();
  });
});
