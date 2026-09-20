import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { hasPermission, PermissionGate } from "@/components/permission-gate";
import { renderWithProviders } from "@/lib/tests";

describe("hasPermission", () => {
  it("requires all codes by default", () => {
    expect(hasPermission(["a", "b"], ["a"])).toBe(false);
    expect(hasPermission(["a", "b"], ["a", "b"])).toBe(true);
  });

  it("supports requireAny", () => {
    expect(hasPermission(["a", "b"], ["a"], false)).toBe(true);
  });
});

describe("PermissionGate", () => {
  it("renders children when allowed", () => {
    renderWithProviders(
      <PermissionGate required="customers.create" granted={["customers.create"]}>
        <button type="button">Add</button>
      </PermissionGate>,
    );
    expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument();
  });

  it("hides children when not allowed", () => {
    const { container } = renderWithProviders(
      <PermissionGate required="invoices.void" granted={["customers.create"]}>
        <button type="button">Void</button>
      </PermissionGate>,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("dims children in disable mode", () => {
    renderWithProviders(
      <PermissionGate required="invoices.void" granted={["customers.create"]} mode="disable">
        <button type="button">Void</button>
      </PermissionGate>,
    );
    const disabled = screen.getByText("Void").closest("[aria-disabled]");
    expect(disabled).toHaveAttribute("aria-disabled", "true");
  });
});
