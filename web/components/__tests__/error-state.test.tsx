import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ErrorState } from "../error-state";

describe("ErrorState", () => {
  it("renders the translated copy and a reset button", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    renderWithProviders(<ErrorState onReset={onReset} />);

    expect(screen.getByRole("heading", { level: 1 })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(onReset).toHaveBeenCalledOnce();
  });

  it("links to support and back home", () => {
    renderWithProviders(<ErrorState onReset={() => {}} />);

    expect(screen.getByRole("link", { name: "Hubungi dukungan" })).toHaveAttribute(
      "href",
      "/support",
    );
    expect(screen.getByRole("link", { name: "Kembali ke beranda" })).toHaveAttribute(
      "href",
      "/login",
    );
  });

  it("shows the message and digest when provided", () => {
    renderWithProviders(<ErrorState onReset={() => {}} message="boom" digest="abc123" />);

    expect(screen.getByText("boom")).toBeInTheDocument();
    expect(screen.getByText(/abc123/)).toBeInTheDocument();
  });

  it("omits the detail block when nothing is provided", () => {
    renderWithProviders(<ErrorState onReset={() => {}} />);

    expect(screen.queryByText("boom")).not.toBeInTheDocument();
  });

  it("shows only the digest when the message is missing", () => {
    renderWithProviders(<ErrorState onReset={() => {}} digest="xyz" />);

    expect(screen.getByText(/xyz/)).toBeInTheDocument();
  });
});
