import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { SalaryRuleFormDialog } from "../salary-rule-form-dialog";

beforeEach(() => {});

describe("SalaryRuleFormDialog", () => {
  it("renders the code and name fields", async () => {
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={[]}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByLabelText("Kode")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("accepts a rule code", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={[]}
        onSave={vi.fn()}
      />,
    );

    const codeInput = await screen.findByLabelText("Kode");
    await user.type(codeInput, "BASIC");

    expect(codeInput).toHaveValue("BASIC");
  });
});
