import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { renderWithProviders } from "@/lib/tests";

type Values = {
  name: string;
};

function Harness({
  onSubmit,
  onOpenChange,
  isPending,
}: {
  onSubmit: (values: Values) => void;
  onOpenChange: (open: boolean) => void;
  isPending?: boolean;
}) {
  const form = useForm<Values>({ defaultValues: { name: "" } });

  return (
    <EntityFormDialog
      open
      onOpenChange={onOpenChange}
      title="New customer"
      description="Add a customer"
      form={form}
      onSubmit={onSubmit}
      isPending={isPending}
    >
      <input aria-label="Name" {...form.register("name")} />
    </EntityFormDialog>
  );
}

describe("regression: EntityFormDialog cancel button must not submit", () => {
  it("should close without submitting when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const handleSubmit = vi.fn();
    const handleOpenChange = vi.fn();

    renderWithProviders(<Harness onSubmit={handleSubmit} onOpenChange={handleOpenChange} />);

    const cancel = screen.getByRole("button", { name: "Cancel" });
    expect(cancel).toHaveAttribute("type", "button");

    await user.click(cancel);

    expect(handleOpenChange).toHaveBeenCalledWith(false);
    expect(handleSubmit).not.toHaveBeenCalled();
  });

  it("should disable both actions while pending", () => {
    const handleSubmit = vi.fn();
    const handleOpenChange = vi.fn();

    renderWithProviders(
      <Harness onSubmit={handleSubmit} onOpenChange={handleOpenChange} isPending />,
    );

    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Saving…" })).toBeDisabled();
  });
});
