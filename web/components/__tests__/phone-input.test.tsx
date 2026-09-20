import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { getDialCodePrefix, PhoneInput } from "@/components/phone-input";
import { renderWithProviders } from "@/lib/tests";

describe("getDialCodePrefix", () => {
  it("detects the dial code from an initial phone value", () => {
    expect(getDialCodePrefix("+62 21 1234 5678")).toBe("62");
    expect(getDialCodePrefix("+65 8123 4567")).toBe("65");
    expect(getDialCodePrefix("+1 (555) 0100")).toBe("1");
    expect(getDialCodePrefix("no number")).toBe("");
    expect(getDialCodePrefix(undefined)).toBe("");
  });
});

describe("PhoneInput", () => {
  it("defaults to Indonesia when the initial value starts with +62", () => {
    renderWithProviders(<PhoneInput id="phone" defaultValue="+62 21 1234 5678" />);

    expect(screen.getByRole("combobox", { name: /Indonesia/ })).toHaveTextContent("+62");
  });

  it("lets the user pick another country's dial code", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PhoneInput defaultValue="+62 21 1234 5678" />);

    await user.click(screen.getByRole("combobox", { name: /Indonesia/ }));
    await user.click(await screen.findByRole("option", { name: /Singapore/ }));

    expect(screen.getByRole("combobox", { name: /Singapore/ })).toHaveTextContent("+65");
  });

  it("keeps the phone number typed in the field independent of the country selection", () => {
    renderWithProviders(<PhoneInput defaultValue="+62 21 1234 5678" />);

    const field = screen.getByPlaceholderText("+123456789");
    fireEvent.change(field, { target: { value: "+1 555 0100" } });

    expect(field).toHaveValue("+1 555 0100");
  });
});
