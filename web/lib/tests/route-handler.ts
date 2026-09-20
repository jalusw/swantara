import { NextRequest } from "next/server";

export type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  url?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

export function buildRequest({
  method = "GET",
  url = "http://localhost/api",
  body,
  headers,
}: RequestOptions = {}): NextRequest {
  return new NextRequest(url, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export type RouteHandlerResult = {
  status: number;
  body: unknown;
  response: Response;
};

export async function callRoute(
  route: (request: NextRequest) => Response | Promise<Response>,
  options: RequestOptions = {},
): Promise<RouteHandlerResult> {
  const response = await route(buildRequest(options));
  let body: unknown = null;
  try {
    body = await response.json();
  } catch {
    // Non-JSON bodies are fine; `body` stays null.
  }
  return { status: response.status, body, response };
}
