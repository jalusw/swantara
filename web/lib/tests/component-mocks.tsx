import { vi } from "vitest";

vi.mock("@/components/button", () => ({
  Button: ({
    children,
    asChild,
    ...props
  }: React.ComponentProps<"button"> & { asChild?: boolean }) =>
    asChild ? (
      children
    ) : (
      <button type="button" {...props}>
        {children}
      </button>
    ),
}));

vi.mock("@/components/separator", () => ({
  Separator: () => <hr />,
}));
