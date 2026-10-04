import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CrmLead } from "@/lib/services/swantara/types";
import { renderWithProviders, server } from "@/lib/tests";
import { PipelineBoard } from "../pipeline-board";

const STAMP = "2026-01-01T00:00:00Z";

function opp(id: number, patch: Partial<CrmLead> = {}): CrmLead {
  return {
    id,
    organizationId: 1,
    name: `Deal ${id}`,
    type: "opportunity",
    contactId: null,
    contactName: null,
    email: null,
    phone: null,
    jobPosition: null,
    stageId: 1,
    expectedRevenue: 20000,
    probability: 30,
    priority: 1,
    salespersonId: null,
    salesGroupId: null,
    source: null,
    medium: null,
    campaign: null,
    lostReason: null,
    expectedClose: null,
    closedAt: null,
    isWon: false,
    createdAt: new Date(STAMP),
    updatedAt: new Date(STAMP),
    ...patch,
  };
}

function stage(id: number, patch: Record<string, unknown> = {}) {
  return {
    id,
    organizationId: 1,
    name: `Stage ${id}`,
    sequence: id,
    isWon: false,
    probability: 10 * id,
    createdAt: new Date(STAMP),
    updatedAt: new Date(STAMP),
    ...patch,
  };
}

const baseStages = [stage(1), stage(2, { name: "Proposal", probability: 60 })];

beforeEach(() => {});

describe("PipelineBoard", () => {
  it("renders open opportunities as cards with revenue subtitles", () => {
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[opp(11), opp(12)]}
        onRefresh={() => {}}
      />,
    );

    expect(screen.getByText("Deal 11")).toBeInTheDocument();
    expect(screen.getByText("Deal 12")).toBeInTheDocument();
    expect(
      screen.getByText("Seret kartu antar tahap, atau gunakan menu pada kartu.", { exact: false }),
    ).toBeInTheDocument();
  });

  it("hides closed and lost opportunities from the board", () => {
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[
          opp(11),
          opp(12, { closedAt: new Date("2026-03-01T00:00:00Z") }),
          opp(13, { lostReason: "Too expensive" }),
        ]}
        onRefresh={() => {}}
      />,
    );

    expect(screen.getByText("Deal 11")).toBeInTheDocument();
    expect(screen.queryByText("Deal 12")).not.toBeInTheDocument();
    expect(screen.queryByText("Deal 13")).not.toBeInTheDocument();
  });

  it("shows won banners with quotation links", () => {
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[opp(11), opp(12, { isWon: true })]}
        onRefresh={() => {}}
      />,
    );

    expect(screen.getByText("Menang")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Buat penawaran" })).toHaveAttribute(
      "href",
      "/crm?won=12",
    );
  });

  it("falls back to the first column for unassigned opportunities", () => {
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[opp(11, { stageId: null })]}
        onRefresh={() => {}}
      />,
    );

    expect(screen.getByText("Deal 11")).toBeInTheDocument();
  });

  it("advances the stage through the card move menu", async () => {
    let advanceCalls = 0;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/crm/opportunities/:id/advance-stage",
        () => {
          advanceCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: { opportunity: {} } });
        },
      ),
    );
    const onRefresh = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[opp(11)]}
        onRefresh={onRefresh}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Move Deal 11" }));
    await user.click(await screen.findByRole("option", { name: /Proposal/ }));

    await waitFor(() => expect(advanceCalls).toBe(1));
  });

  it("wins through the card move menu when the target is a won stage", async () => {
    let winCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/crm/opportunities/:id/win", () => {
        winCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { opportunity: {} } });
      }),
    );
    const wonStages = [stage(1), stage(9, { name: "Won", isWon: true, probability: 100 })];
    const user = userEvent.setup();
    renderWithProviders(
      <PipelineBoard orgId="1" stages={wonStages} opportunities={[opp(11)]} onRefresh={() => {}} />,
    );

    await user.click(screen.getByRole("combobox", { name: "Move Deal 11" }));
    await user.click(await screen.findByRole("option", { name: /Won/ }));

    await waitFor(() => expect(winCalls).toBe(1));
  });

  it("opens the lost dialog when the target is a lost stage", async () => {
    const lostStages = [stage(1), stage(9, { name: "Lost", probability: 0 })];
    const user = userEvent.setup();
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={lostStages}
        opportunities={[opp(11)]}
        onRefresh={() => {}}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Move Deal 11" }));
    await user.click(await screen.findByRole("option", { name: /Lost/ }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("wins from the mobile quick action", async () => {
    let winCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/crm/opportunities/:id/win", () => {
        winCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { opportunity: {} } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(
      <PipelineBoard
        orgId="1"
        stages={baseStages}
        opportunities={[opp(11)]}
        onRefresh={() => {}}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Tandai menang: Deal 11" }));

    await waitFor(() => expect(winCalls).toBe(1));
  });
});
