import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeFormDialog } from "../employee-form-dialog";

function useLocalRefs() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { job_positions: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalRefs();
});

describe("EmployeeFormDialog", () => {
  it("renders create title", async () => {
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Tambah Karyawan")).toBeInTheDocument();
  });

  it("renders employee number field", async () => {
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Tambah Karyawan");
    expect(screen.getByLabelText("Nomor karyawan")).toBeInTheDocument();
  });
});
