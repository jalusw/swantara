import { screen } from "@testing-library/react";
import { act } from "react";
import { describe, expect, it, vi } from "vitest";
import { InfiniteList, type InfiniteListProps } from "@/components/infinite-list";
import { intersectionObserverMock, renderWithProviders } from "@/lib/tests";

type Item = { id: string; label: string };

function renderList(props: Partial<InfiniteListProps<Item>> = {}) {
  const items: Item[] = [
    { id: "a", label: "Alpha" },
    { id: "b", label: "Beta" },
  ];
  const onEndReached = vi.fn();

  renderWithProviders(
    <InfiniteList
      items={items}
      renderItem={(item) => <p>{item.label}</p>}
      getItemKey={(item) => item.id}
      onEndReached={onEndReached}
      {...props}
    />,
  );

  return { onEndReached };
}

describe("InfiniteList", () => {
  it("renders every item in an accessible feed", () => {
    renderList();

    expect(screen.getByRole("feed", { name: "Infinite list" })).toBeInTheDocument();
    expect(screen.getByText("Alpha")).toBeInTheDocument();
    expect(screen.getByText("Beta")).toBeInTheDocument();
  });

  it("calls onEndReached when the end scrolls into view", () => {
    const { onEndReached } = renderList();

    act(() => intersectionObserverMock.triggerAll(true));

    expect(onEndReached).toHaveBeenCalledOnce();
  });

  it("does not call onEndReached while loading", () => {
    const { onEndReached } = renderList({ loading: true });

    act(() => intersectionObserverMock.triggerAll(true));

    expect(onEndReached).not.toHaveBeenCalled();
  });

  it("does not observe or load more when hasMore is false", () => {
    const { onEndReached } = renderList({ hasMore: false });

    act(() => intersectionObserverMock.triggerAll(true));

    expect(onEndReached).not.toHaveBeenCalled();
    expect(document.querySelector('[data-slot="infinite-list-sentinel"]')).not.toBeInTheDocument();
  });

  it("shows the empty state when there are no items", () => {
    renderList({
      items: [],
      emptyState: <p>Nothing here</p>,
    });

    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });

  it("shows a custom loading indicator while loading", () => {
    renderList({
      items: [],
      loading: true,
      loadingIndicator: <p>Fetching…</p>,
    });

    expect(screen.getByText("Fetching…")).toBeInTheDocument();
  });

  it("renders a default loading indicator when none is provided", () => {
    renderList({ items: [], loading: true });

    expect(screen.getByRole("status")).toBeInTheDocument();
  });
});
