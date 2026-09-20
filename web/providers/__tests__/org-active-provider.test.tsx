import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useOrganizationId } from "@/lib/hooks/use-org-context";
import { renderWithProviders } from "@/lib/tests";
import { OrgActiveContext, OrgActiveProvider } from "../org-active-provider";

describe("OrgActiveProvider", () => {
  it("provides the route organization id to children", () => {
    render(
      <OrgActiveProvider orgId={42}>
        <OrgActiveContext.Consumer>
          {(value) => <p>{value === null ? "empty" : `org-${value}`}</p>}
        </OrgActiveContext.Consumer>
      </OrgActiveProvider>,
    );
    expect(screen.getByText("org-42")).toBeInTheDocument();
  });

  it("prefers the route id over the store id", () => {
    function Probe() {
      return <p>{`org-${useOrganizationId()}`}</p>;
    }
    renderWithProviders(
      <OrgActiveProvider orgId={9}>
        <Probe />
      </OrgActiveProvider>,
    );
    expect(screen.getByText("org-9")).toBeInTheDocument();
  });
});
