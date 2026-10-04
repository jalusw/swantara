import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SalaryRuleFormDialog } from "../salary-rule-form-dialog";

const accounts = [
  { id: 11, code: "5000", name: "Salary Expense" },
  { id: 12, code: "2000", name: "Payable" },
];

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/salary-rules", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { salary_rule: { id: 9 } },
      });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/salary-rules/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { salary_rule: { id: 4 } },
      });
    }),
  );
}

const editInitial = {
  id: 4,
  code: "BASIC",
  name: "Basic Salary",
  category: "earning",
  computeType: "fixed",
  amount: null,
  formula: null,
  accountDebitId: null,
  accountCreditId: null,
} as never;

const editInitialFilled = {
  id: 5,
  code: "TAX",
  name: "Income Tax",
  category: "deduction",
  computeType: "formula",
  amount: 10,
  formula: "base * 0.1",
  accountDebitId: 11,
  accountCreditId: 12,
} as never;

const editInitialFallback = {
  id: 6,
  code: "ALLOW",
  name: "Allowance",
  category: null,
  computeType: null,
  amount: null,
  formula: null,
  accountDebitId: null,
  accountCreditId: null,
} as never;

beforeEach(() => {});

describe("SalaryRuleFormDialog branches", () => {
  it("shows edit title with prefilled values", async () => {
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        accounts={accounts as never}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Ubah aturan gaji")).toBeInTheDocument();
    expect(screen.getByDisplayValue("base * 0.1")).toBeInTheDocument();
  });

  it("falls back to earning and fixed when initial lacks them", async () => {
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFallback}
        accounts={accounts as never}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Ubah aturan gaji")).toBeInTheDocument();
    expect(screen.getByDisplayValue("ALLOW")).toBeInTheDocument();
  });

  it("blocks submit when code and name are empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await screen.findByLabelText("Kode");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).not.toHaveBeenCalled());
  });

  it("creates a rule with only code and name", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Kode"), "BASIC");
    await user.type(screen.getByLabelText("Nama"), "Basic Salary");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates a deduction rule with formula and accounts", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Kode"), "TAX");
    await user.type(screen.getByLabelText("Nama"), "Income Tax");
    await user.click(screen.getByRole("combobox", { name: "Kategori" }));
    await user.click(
      await screen.findByRole("option", {
        name: "Potongan",
      }),
    );
    await user.click(screen.getByRole("combobox", { name: "Jenis perhitungan" }));
    await user.click(
      await screen.findByRole("option", {
        name: "Formula",
      }),
    );
    await user.type(screen.getByLabelText("Amount"), "10");
    await user.type(screen.getByLabelText("Formula"), "base * 0.1");
    await user.click(screen.getByRole("combobox", { name: "Akun debit" }));
    await user.click(await screen.findByRole("option", { name: "5000 Salary Expense" }));
    await user.click(screen.getByRole("combobox", { name: "Akun kredit" }));
    await user.click(await screen.findByRole("option", { name: "2000 Payable" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Kode"), "BASIC");
    await user.type(screen.getByLabelText("Nama"), "Basic Salary");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByLabelText("Kode")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a rule with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah aturan gaji");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates a rule with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah aturan gaji");
    await user.clear(screen.getByDisplayValue("base * 0.1"));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        accounts={accounts as never}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah aturan gaji");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Ubah aturan gaji")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <SalaryRuleFormDialog
        open={true}
        onOpenChange={onOpenChange}
        orgId="1"
        initial={null}
        accounts={accounts as never}
        onSave={vi.fn()}
      />,
    );

    await screen.findByLabelText("Kode");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
