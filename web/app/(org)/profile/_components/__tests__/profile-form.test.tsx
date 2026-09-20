import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ProfileForm } from "../profile-form";

describe("ProfileForm", () => {
  it("renders the signed-in user profile for editing", async () => {
    renderWithProviders(<ProfileForm />);

    expect(await screen.findByDisplayValue("Alex")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Rivera")).toBeInTheDocument();
  });

  it("accepts edits to the first name field", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    const firstName = await screen.findByDisplayValue("Alex");
    await user.clear(firstName);
    await user.type(firstName, "Alexander");

    expect(firstName).toHaveValue("Alexander");
  });
});
