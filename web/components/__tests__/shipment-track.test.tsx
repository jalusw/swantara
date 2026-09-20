import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { type ShipmentStep, ShipmentTrack } from "@/components/shipment-track";
import { renderWithProviders } from "@/lib/tests";

const STEPS: ShipmentStep[] = [
  {
    id: "placed",
    label: "Order placed",
    description: "Acme Inc",
    timestamp: "Jan 2",
    state: "done",
  },
  { id: "packed", label: "Packed", timestamp: "Jan 3", state: "done" },
  { id: "transit", label: "In transit", timestamp: "Jan 4", state: "active" },
  { id: "delivered", label: "Delivered", state: "pending" },
];

describe("ShipmentTrack", () => {
  it("renders every step with labels and timestamps", () => {
    renderWithProviders(<ShipmentTrack steps={STEPS} />);

    expect(screen.getByText("Order placed")).toBeInTheDocument();
    expect(screen.getByText("Delivered")).toBeInTheDocument();
    expect(screen.getByText("Jan 2")).toBeInTheDocument();
    expect(screen.getByText("In transit")).toBeInTheDocument();
  });

  it("fills the progress bar to the active step", () => {
    const { container } = renderWithProviders(<ShipmentTrack steps={STEPS} />);

    const progress = container.querySelector('[data-slot="shipment-track-progress"]');
    expect(progress).toHaveStyle({ width: `${(2 / 3) * 100}%` });
  });

  it("marks completed steps with a check icon only", () => {
    const { container } = renderWithProviders(<ShipmentTrack steps={STEPS} />);

    expect(container.querySelectorAll('[data-slot="shipment-track"] .lucide-check')).toHaveLength(
      2,
    );
  });

  it("reaches 100% when every step is done", () => {
    const { container } = renderWithProviders(
      <ShipmentTrack steps={STEPS.map((step) => ({ ...step, state: "done" }))} />,
    );

    const progress = container.querySelector('[data-slot="shipment-track-progress"]');
    expect(progress).toHaveStyle({ width: "100%" });
  });
});
