import { zodResolver } from "@hookform/resolvers/zod";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { z } from "zod";
import { Form, FormField, SubmitButton } from "@/components/form";
import { renderWithProviders } from "@/lib/tests";

const schema = z.object({
  name: z.string().min(2, "Name is too short"),
});
type Values = z.infer<typeof schema>;

function DemoForm({ onSubmit }: { onSubmit: (values: Values) => void }) {
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { name: "" },
  });

  return (
    <Form form={form} onSubmit={onSubmit}>
      <FormField name="name" label="Name" description="Display name.">
        {({ field, fieldState, id }) => (
          <input
            id={id}
            placeholder="Full name"
            aria-invalid={fieldState.error ? true : undefined}
            aria-describedby={fieldState.error ? `${id}-feedback` : undefined}
            {...field}
          />
        )}
      </FormField>
      <SubmitButton>Save</SubmitButton>
    </Form>
  );
}

describe("Form + FormField + SubmitButton", () => {
  it("renders label, description, and submit control", () => {
    renderWithProviders(<DemoForm onSubmit={vi.fn()} />);
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Display name.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
  });

  it("surfaces validation errors", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DemoForm onSubmit={vi.fn()} />);
    await user.type(screen.getByPlaceholderText("Full name"), "a");
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Name is too short")).toBeInTheDocument();
  });

  it("submits valid values", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    renderWithProviders(<DemoForm onSubmit={onSubmit} />);
    await user.type(screen.getByPlaceholderText("Full name"), "Acme Inc.");
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit.mock.calls[0]![0]).toEqual({ name: "Acme Inc." });
  });
});
