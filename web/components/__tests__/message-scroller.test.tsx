import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  MessageScroller,
  MessageScrollerContent,
  MessageScrollerItem,
  MessageScrollerProvider,
  MessageScrollerViewport,
} from "@/components/message-scroller";
import { renderWithProviders } from "@/lib/tests";

function renderScroller(children: React.ReactNode) {
  return renderWithProviders(
    <MessageScrollerProvider>
      <MessageScroller>
        <MessageScrollerViewport>
          <MessageScrollerContent>{children}</MessageScrollerContent>
        </MessageScrollerViewport>
      </MessageScroller>
    </MessageScrollerProvider>,
  );
}

describe("MessageScroller", () => {
  it("renders with data-slot", () => {
    renderScroller(<MessageScrollerItem>Message 1</MessageScrollerItem>);

    expect(
      screen.getByText("Message 1").closest("[data-slot='message-scroller']"),
    ).toBeInTheDocument();
  });

  it("renders viewport with default aria-label", () => {
    renderScroller(<></>);

    expect(screen.getByRole("region", { name: "Messages" })).toBeInTheDocument();
  });

  it("renders items inside content", () => {
    renderScroller(
      <>
        <MessageScrollerItem>First</MessageScrollerItem>
        <MessageScrollerItem>Second</MessageScrollerItem>
      </>,
    );

    expect(screen.getByText("First")).toHaveAttribute("data-slot", "message-scroller-item");
    expect(screen.getByText("Second")).toHaveAttribute("data-slot", "message-scroller-item");
  });
});
