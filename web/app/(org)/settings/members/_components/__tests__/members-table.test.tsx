import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { type MemberRow, MembersTable } from "../members-table";

const rows: MemberRow[] = [
  {
    id: "1",
    name: "Alex Rivera",
    email: "alex@acme.com",
    role: "owner",
    status: "active",
    addedAt: "Jan 5, 2026",
  },
  {
    id: "2",
    name: "June Park",
    email: "june@acme.com",
    role: "admin",
    status: "active",
    addedAt: "Feb 10, 2026",
  },
  {
    id: "3",
    name: "—",
    email: "—",
    role: "member",
    status: "invited",
    addedAt: "Mar 1, 2026",
  },
];

describe("MembersTable", () => {
  it("renders rows with role and status badges", () => {
    renderWithProviders(<MembersTable members={rows} />);

    expect(screen.getByText("Alex Rivera")).toBeInTheDocument();
    expect(screen.getByText("June Park")).toBeInTheDocument();
    expect(screen.getByText("Pemilik")).toBeInTheDocument();
    expect(screen.getByText("Admin")).toBeInTheDocument();
    expect(screen.getAllByText("Aktif")).toHaveLength(2);
    expect(screen.getByText("Diundang")).toBeInTheDocument();
  });

  it("filters rows by search query", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MembersTable members={rows} />);

    await user.type(screen.getByPlaceholderText("Cari anggota…"), "June");

    expect(screen.getByText("June Park")).toBeInTheDocument();
    expect(screen.queryByText("Alex Rivera")).toBeNull();
  });
});
