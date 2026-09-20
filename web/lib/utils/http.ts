export const jsonContentType = "application/json";

export function isJson(contentType: string | null): boolean {
  return !!contentType && contentType.split(";")[0]!.trim().toLowerCase() === jsonContentType;
}
