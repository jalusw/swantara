import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { WorkflowSteps } from "@/components/workflow-steps";
import { renderWithProviders } from "@/lib/tests";

describe("WorkflowSteps", () => {
  it("marks the current step", () => {
    renderWithProviders(
      <WorkflowSteps steps={[{ label: "A" }, { label: "B" }, { label: "C" }]} currentIndex={1} />,
    );
    const stepButton = screen.getByText("B").closest("button") as HTMLElement;
    expect(stepButton).toHaveAttribute("aria-current", "step");
  });

  it("calls onStepClick for a clicked step", async () => {
    const user = userEvent.setup();
    const onStepClick = vi.fn();
    renderWithProviders(
      <WorkflowSteps
        steps={[{ label: "A" }, { label: "B" }, { label: "C" }]}
        currentIndex={0}
        onStepClick={onStepClick}
      />,
    );
    await user.click(screen.getByText("C"));
    expect(onStepClick).toHaveBeenCalledWith(2);
  });
});
