import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProfileForm } from "../profile-form";

beforeEach(() => {});

describe("ProfileForm extra", () => {
  it("shows the error state when the profile fails to load", async () => {
    server.use(
      http.get("*/api/v1/me", () =>
        HttpResponse.json({ success: false, message: "Unavailable." }, { status: 500 }),
      ),
    );
    renderWithProviders(<ProfileForm />);

    expect(await screen.findByRole("button", { name: "Try again" })).toBeInTheDocument();
  });

  it("retries loading the profile from the error state", async () => {
    let meCalls = 0;
    server.use(
      http.get("*/api/v1/me", () => {
        meCalls += 1;
        if (meCalls === 1) {
          return HttpResponse.json({ success: false, message: "Unavailable." }, { status: 500 });
        }
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            user: {
              id: 1,
              username: "alex",
              first_name: "Alex",
              last_name: "Rivera",
              email: "alex@acme.com",
              phone: null,
              avatar: null,
              bio: null,
              birthday: null,
              active: true,
              sex: null,
              address: null,
              city: null,
              postal_code: null,
              system_role_id: 1,
              email_verified_at: null,
              phone_verified_at: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          },
        });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    await user.click(await screen.findByRole("button", { name: "Try again" }));

    expect(await screen.findByDisplayValue("Alex")).toBeInTheDocument();
  });

  it("saves profile changes", async () => {
    const updateCalls: unknown[] = [];
    server.use(
      http.put("*/api/v1/users/:id", async ({ request }) => {
        updateCalls.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    const bio = await screen.findByPlaceholderText("Tell us a little about yourself");
    await user.type(bio, "Operations lead");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    await waitFor(() => expect(updateCalls).toHaveLength(1));
    expect(updateCalls[0]).toMatchObject({ bio: "Operations lead" });
  });

  it("offers avatar upload and sex selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    await screen.findByDisplayValue("Alex");
    expect(screen.getByRole("button", { name: "Upload photo" })).toBeInTheDocument();
    await user.click(screen.getByRole("combobox", { name: "Sex" }));
    expect(await screen.findByRole("option", { name: "Prefer not to say" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Male" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Female" })).toBeInTheDocument();
  });
});
