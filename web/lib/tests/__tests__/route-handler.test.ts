// @vitest-environment node
import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { buildRequest, callRoute } from "../route-handler";

describe("buildRequest", () => {
  it("builds a default GET request", () => {
    const request = buildRequest();
    expect(request).toBeInstanceOf(NextRequest);
    expect(request.method).toBe("GET");
  });

  it("serializes a JSON body with merged headers", () => {
    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/items",
      body: { name: "Acme" },
      headers: { "x-requested-with": "Swantara" },
    });
    expect(request.method).toBe("POST");
    expect(request.headers.get("x-requested-with")).toBe("Swantara");
  });

  it("omits the body when none is given", async () => {
    const request = buildRequest({ method: "GET", url: "http://localhost/api/items" });
    expect(await request.text()).toBe("");
  });
});

describe("callRoute", () => {
  it("returns the status and parsed body", async () => {
    const result = await callRoute(
      () => Response.json({ success: true, message: "ok" }, { status: 201 }),
      { method: "GET" },
    );
    expect(result.status).toBe(201);
    expect(result.body).toEqual({ success: true, message: "ok" });
  });

  it("keeps a null body for non-JSON responses", async () => {
    const result = await callRoute(() => new Response("plain", { status: 200 }));
    expect(result.status).toBe(200);
    expect(result.body).toBeNull();
  });
});
