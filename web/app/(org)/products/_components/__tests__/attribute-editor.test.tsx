import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { AttributeEditor, type AttributeRow } from "../attribute-editor";

function renderEditor(initial: AttributeRow[]) {
  function Harness() {
    const [value, setValue] = useState<AttributeRow[]>(initial);
    return <AttributeEditor value={value} onChange={setValue} />;
  }
  return renderWithProviders(<Harness />);
}

describe("AttributeEditor", () => {
  it("renders existing attributes with values", () => {
    renderEditor([{ id: "a1", name: "Color", values: ["Red", "Blue"] }]);

    expect(screen.getByDisplayValue("Color")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Red, Blue")).toBeInTheDocument();
  });

  it("adds a new attribute row", async () => {
    const user = userEvent.setup();
    renderEditor([]);

    await user.click(screen.getByRole("button", { name: "Add attribute" }));

    expect(screen.getByPlaceholderText("Attribute")).toBeInTheDocument();
  });

  it("edits attribute values through the inputs", async () => {
    const user = userEvent.setup();
    renderEditor([{ id: "a1", name: "Size", values: [] }]);

    await user.type(screen.getByDisplayValue(""), "Blue");

    expect(screen.getByDisplayValue("Blue")).toBeInTheDocument();
  });

  it("removes an attribute row", async () => {
    const user = userEvent.setup();
    renderEditor([{ id: "a1", name: "Color", values: ["Red"] }]);

    await user.click(screen.getByRole("button", { name: "Remove" }));

    expect(screen.queryByDisplayValue("Color")).not.toBeInTheDocument();
  });
});
