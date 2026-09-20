import { cn } from "@/lib/utils";

export type ContainerProps = {
  children?: React.ReactNode;
  className?: string;
};

function Container({ children, className }: ContainerProps) {
  return (
    <div
      data-slot="container"
      className={cn("mx-auto w-full max-w-screen-xl px-4 sm:px-6 lg:px-8", className)}
    >
      {children}
    </div>
  );
}

export { Container };
