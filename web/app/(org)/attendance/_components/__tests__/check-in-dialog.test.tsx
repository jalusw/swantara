import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Employee } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { CheckInDialog } from "../check-in-dialog";

const employees = [
  {
    id: 1,
    organizationId: 1,
    contactId: 10,
    userId: null,
    employeeNumber: "EMP-0001",
    departmentId: null,
    jobPositionId: null,
    managerId: null,
    hireDate: "2024-01-01",
    terminationDate: null,
    employmentType: "full_time",
    workLocation: null,
    active: true,
  },
] as Employee[];

beforeEach(() => {});

describe("CheckInDialog", () => {
  it("renders the employee selector", async () => {
    renderWithProviders(
      <CheckInDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        employees={employees}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Karyawan" })).toBeInTheDocument();
  });

  it("closes when the cancel action is used", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <CheckInDialog
        open={true}
        onOpenChange={onOpenChange}
        orgId="1"
        employees={employees}
        onSave={vi.fn()}
      />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: /batal/i }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
