import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { AttachmentList } from "@/components/attachment-list";
import { renderWithProviders } from "@/lib/tests";

const attachments = [
  {
    id: "a1",
    name: "invoice.pdf",
    size: 24576,
    downloadUrl: "/files/invoice.pdf",
  },
  { id: "a2", name: "photo.png", size: 123456 },
  { id: "a3", name: "ledger.csv", size: 2048 },
];

describe("AttachmentList", () => {
  it("renders each attachment with size", () => {
    renderWithProviders(<AttachmentList attachments={attachments} />);

    expect(screen.getByText("invoice.pdf")).toBeInTheDocument();
    expect(screen.getAllByText("24 KB").length).toBe(1);
    expect(screen.getByText("120.6 KB")).toBeInTheDocument();
  });

  it("removes an attachment via the remove action", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    renderWithProviders(<AttachmentList attachments={attachments} onRemove={onRemove} />);

    await user.click(screen.getByRole("button", { name: "Remove photo.png" }));
    expect(onRemove).toHaveBeenCalledWith("a2");
  });

  it("links downloads when a url is provided", () => {
    renderWithProviders(<AttachmentList attachments={attachments} />);
    const link = screen.getByRole("link", { name: "Download invoice.pdf" });
    expect(link).toHaveAttribute("href", "/files/invoice.pdf");
    expect(link).toHaveAttribute("download");
  });
});
