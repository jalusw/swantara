import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Attachment,
  AttachmentContent,
  AttachmentDescription,
  AttachmentMedia,
  AttachmentTitle,
} from "@/components/attachment";
import { renderWithProviders } from "@/lib/tests";

describe("Attachment", () => {
  it("renders with default data-slot and state", () => {
    renderWithProviders(
      <Attachment>
        <AttachmentContent>
          <AttachmentTitle>document.pdf</AttachmentTitle>
        </AttachmentContent>
      </Attachment>,
    );

    const el = screen.getByText("document.pdf").closest("[data-slot='attachment']");
    expect(el).toHaveAttribute("data-state", "done");
    expect(el).toHaveAttribute("data-size", "default");
  });

  it("renders title and description inside content", () => {
    renderWithProviders(
      <Attachment>
        <AttachmentContent>
          <AttachmentTitle>photo.png</AttachmentTitle>
          <AttachmentDescription>2.4 MB</AttachmentDescription>
        </AttachmentContent>
      </Attachment>,
    );

    expect(screen.getByText("photo.png")).toBeInTheDocument();
    expect(screen.getByText("2.4 MB")).toBeInTheDocument();
  });

  it("renders media slot with data-variant", () => {
    renderWithProviders(
      <Attachment>
        <AttachmentMedia variant="icon" />
        <AttachmentContent>
          <AttachmentTitle>file.txt</AttachmentTitle>
        </AttachmentContent>
      </Attachment>,
    );

    const media = screen
      .getByText("file.txt")
      .closest("[data-slot='attachment']")
      ?.querySelector("[data-slot='attachment-media']");
    expect(media).toHaveAttribute("data-variant", "icon");
  });
});
