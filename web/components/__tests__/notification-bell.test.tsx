import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { NotificationBell } from "@/components/notification-bell";
import { renderWithProviders } from "@/lib/tests";

const items = [
  { id: "1", title: "Invoice approved", unread: true },
  { id: "2", title: "New message", unread: false },
];

describe("NotificationBell", () => {
  it("shows the count of unread items", () => {
    renderWithProviders(<NotificationBell items={items} />);
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("uses an explicit unreadCount when provided", () => {
    renderWithProviders(<NotificationBell items={items} unreadCount={5} />);
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("calls onRead when an item is opened and clicked", async () => {
    const user = userEvent.setup();
    const onRead = vi.fn();
    renderWithProviders(<NotificationBell items={items} onRead={onRead} />);
    await user.click(screen.getByRole("button", { name: /unread/ }));
    const item = await screen.findByRole("button", {
      name: /Invoice approved/,
    });
    await user.click(item);
    expect(onRead).toHaveBeenCalledWith("1");
  });

  it("renders a tooltip when tooltip prop is provided", async () => {
    const user = userEvent.setup();
    renderWithProviders(<NotificationBell items={items} tooltip="Notifications" />);
    await user.hover(screen.getByRole("button", { name: /unread/ }));
    expect(await screen.findByText("Notifications")).toBeInTheDocument();
  });
});
