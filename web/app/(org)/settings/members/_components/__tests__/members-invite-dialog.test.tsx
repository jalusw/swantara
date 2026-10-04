import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MembersInviteDialog } from "../members-invite-dialog";

const users = [
  {
    id: 9,
    username: "janedoe",
    first_name: "Jane",
    last_name: "Doe",
    avatar: null,
    bio: null,
  },
];

function mockSearch() {
  server.use(
    http.get("*/api/v1/users", ({ request }) => {
      const url = new URL(request.url);
      const q = url.searchParams.get("q") ?? "";
      if (q.includes("@")) return HttpResponse.json({ success: true, data: { users: [] } });
      return HttpResponse.json({ success: true, data: { users } });
    }),
  );
}

describe("MembersInviteDialog", () => {
  it("searches by username and invites the selected user", async () => {
    mockSearch();
    let posted: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/members", async ({ request }) => {
        posted = await request.json();
        return HttpResponse.json({ success: true, data: { member: { id: 5 } } }, { status: 201 });
      }),
    );
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<MembersInviteDialog open onOpenChange={onOpenChange} />);

    await user.type(screen.getByLabelText("Cari pengguna"), "jane");
    expect(await screen.findByText("janedoe", { exact: false })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Jane Doe/ }));
    await user.click(screen.getByRole("button", { name: "Undang" }));

    await waitFor(() => expect(posted).toMatchObject({ user_id: 9 }));
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });

  it("shows no suggestions for email input", async () => {
    mockSearch();
    const user = userEvent.setup();
    renderWithProviders(<MembersInviteDialog open onOpenChange={vi.fn()} />);

    await user.type(screen.getByLabelText("Cari pengguna"), "jane@example.com");

    expect(
      await screen.findByText(
        "Tidak ada saran untuk email atau telepon. Cari berdasarkan nama pengguna atau nama.",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText("janedoe", { exact: false })).toBeNull();
  });
});
