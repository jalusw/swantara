import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ActivityTimeline } from "@/components/activity-timeline";
import { renderWithProviders } from "@/lib/tests";

describe("ActivityTimeline", () => {
  it("renders entries with titles and timestamps", () => {
    renderWithProviders(
      <ActivityTimeline
        items={[
          { id: "1", title: "Created", timestamp: "2d ago" },
          { id: "2", title: "Approved", timestamp: "1d ago" },
        ]}
      />,
    );

    expect(screen.getByText("Created")).toBeInTheDocument();
    expect(screen.getByText("Approved")).toBeInTheDocument();
    expect(screen.getByText("2d ago")).toBeInTheDocument();
  });

  it("draws connectors between, not after, the last item", () => {
    const { container } = renderWithProviders(
      <ActivityTimeline
        items={[
          { id: "1", title: "A" },
          { id: "2", title: "B" },
        ]}
      />,
    );
    expect(container.querySelectorAll('[data-slot="activity-timeline-line"]')).toHaveLength(1);
  });
});
