import { zodResolver as baseZodResolver } from "@hookform/resolvers/zod";
import type { FieldValues, Resolver } from "react-hook-form";

export function zodResolver<T extends FieldValues>(
  schema: Parameters<typeof baseZodResolver>[0],
): Resolver<T> {
  return baseZodResolver(schema) as Resolver<T>;
}
