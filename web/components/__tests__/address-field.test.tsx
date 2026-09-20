import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { renderWithProviders } from "@/lib/tests";
import { AddressField, type AddressValue } from "../address-field";

const empty: AddressValue = {
  address: "",
  city: "",
  postalCode: "",
};

function StatefulAddressField({
  initial = empty,
  onChange,
}: {
  initial?: AddressValue;
  onChange: (value: AddressValue) => void;
}) {
  const [value, setValue] = useState(initial);
  return (
    <AddressField
      value={value}
      onChange={(next) => {
        setValue(next);
        onChange(next);
      }}
    />
  );
}

describe("AddressField", () => {
  it("renders labeled inputs for the address fields", () => {
    renderWithProviders(<AddressField value={empty} onChange={() => {}} />);

    expect(screen.getByLabelText("Street address")).toBeInTheDocument();
    expect(screen.getByLabelText("City")).toBeInTheDocument();
    expect(screen.getByLabelText("Postal code")).toBeInTheDocument();
  });

  it("reports text edits through onChange", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<StatefulAddressField onChange={onChange} />);

    await user.type(screen.getByLabelText("City"), "Jakarta");

    expect(onChange).toHaveBeenLastCalledWith({ ...empty, city: "Jakarta" });
  });

  it("disables the inputs while disabled", () => {
    renderWithProviders(<AddressField value={empty} onChange={() => {}} disabled />);

    expect(screen.getByLabelText("Street address")).toBeDisabled();
  });
});
