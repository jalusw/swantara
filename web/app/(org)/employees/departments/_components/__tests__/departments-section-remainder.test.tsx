import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DepartmentsSection } from "../departments-section";

const departments = [
  {
    id: 1,
    organization_id: 1,
    name: "Engineering",
    description: "Item engineering",
    parent_id: null,
    manager_id: null,
    dimension_id: null,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Marketing",
    description: null,
    parent_id: null,
    manager_id: null,
    dimension_id: null,
  },
];

function useDepartmentsHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments } }),
    ),
  );
}

beforeEach(() => {
  useDepartmentsHandlers();
});

describe("DepartmentsSection remainder", () => {
  it("renders a dash for departments without a description", async () => {
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Engineering");
    expect(screen.getByText("Item engineering")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/departments", () =>
        HttpResponse.json({ success: false, message: "Department load failed." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<DepartmentsSection orgId="1" />);

    expect(await screen.findByText("Department load failed.")).toBeInTheDocument();

    useDepartmentsHandlers();
    await user.click(screen.getByRole("button", { name: /coba lagi/i }));

    expect(await screen.findByText("Engineering")).toBeInTheDocument();
  });

  it("opens the edit dialog from the row menu", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Engineering");
    const rowMenus = screen.getAllByRole("button").filter((button) => !button.textContent);
    await user.click(rowMenus[0]!);
    await user.click(await screen.findByRole("menuitem", { name: /ubah/i }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("deletes a department after confirming", async () => {
    let deletedId: number | null = null;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/departments/:id", ({ params }) => {
        deletedId = Number(params.id);
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Marketing");
    const rowMenus = screen.getAllByRole("button").filter((button) => !button.textContent);
    await user.click(rowMenus[1]!);
    await user.click(await screen.findByRole("menuitem", { name: /hapus/i }));

    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /konfirmasi/i, hidden: false }));

    await waitFor(() => expect(deletedId).toBe(2));
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Engineering");
    await user.click(screen.getByRole("button", { name: /tambah departemen/i }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("does nothing when confirming without a selected department", async () => {
    const deleteSpy = vi.fn();
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/departments/:id", () => {
        deleteSpy();
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Engineering");
    expect(deleteSpy).not.toHaveBeenCalled();
  });
});
