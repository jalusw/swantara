const publicPagePatterns = ["/login", "/register", "/forgot-password", "/reset-password"];

export function isProtectedPage(pathname: string): boolean {
  const path = pathname === "" ? "/" : pathname;
  if (path === "/") return false;
  return !publicPagePatterns.some((pattern) => path === pattern || path.startsWith(`${pattern}/`));
}
