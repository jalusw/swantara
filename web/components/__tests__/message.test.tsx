import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Message,
  MessageContent,
  MessageFooter,
  MessageGroup,
  MessageHeader,
} from "@/components/message";
import { renderWithProviders } from "@/lib/tests";

describe("Message", () => {
  it("renders with default align start", () => {
    renderWithProviders(
      <Message>
        <MessageContent>Hello</MessageContent>
      </Message>,
    );

    expect(screen.getByText("Hello").closest("[data-slot='message']")).toHaveAttribute(
      "data-align",
      "start",
    );
  });

  it("renders with align end", () => {
    renderWithProviders(
      <Message align="end">
        <MessageContent>Reply</MessageContent>
      </Message>,
    );

    expect(screen.getByText("Reply").closest("[data-slot='message']")).toHaveAttribute(
      "data-align",
      "end",
    );
  });

  it("renders group, header, and footer slots", () => {
    renderWithProviders(
      <MessageGroup>
        <Message>
          <MessageHeader>Sender</MessageHeader>
          <MessageContent>Body</MessageContent>
          <MessageFooter>Timestamp</MessageFooter>
        </Message>
      </MessageGroup>,
    );

    expect(screen.getByText("Sender")).toHaveAttribute("data-slot", "message-header");
    expect(screen.getByText("Timestamp")).toHaveAttribute("data-slot", "message-footer");
  });
});
