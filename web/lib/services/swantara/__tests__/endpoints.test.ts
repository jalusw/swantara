import { describe, expect, it } from "vitest";
import { endpoints } from "../endpoints";

describe("endpoints", () => {
  it("exposes auth endpoints", () => {
    expect(endpoints.auth.login).toBe("/api/v1/auth/login");
    expect(endpoints.auth.register).toBe("/api/v1/auth/register");
    expect(endpoints.auth.refresh).toBe("/api/v1/auth/refresh");
  });

  it("builds parameterized contact endpoints correctly", () => {
    expect(endpoints.contacts.list("1")).toBe("/api/v1/organizations/1/contacts");
    expect(endpoints.contacts.get("1", "2")).toBe("/api/v1/organizations/1/contacts/2");
  });

  it("builds parameterized item category endpoints correctly", () => {
    expect(endpoints.itemCategories.list("1")).toBe("/api/v1/organizations/1/item-categories");
    expect(endpoints.itemCategories.get("1", "2")).toBe(
      "/api/v1/organizations/1/item-categories/2",
    );
  });

  it("exports all required top-level endpoint groups", () => {
    expect(endpoints.health).toBeDefined();
    expect(endpoints.auth).toBeDefined();
    expect(endpoints.contacts).toBeDefined();
    expect(endpoints.products).toBeDefined();
    expect(endpoints.saleOrders).toBeDefined();
    expect(endpoints.kpis).toBeDefined();
  });

  it("builds every parameterized endpoint without throwing", () => {
    function walk(node: unknown, path: string) {
      if (typeof node === "string") {
        expect(node, path).toContain("/api/v1/");
        return;
      }
      if (typeof node === "function") {
        const built = (node as (...args: string[]) => unknown)("1", "2", "3", "4", "5");
        expect(String(built), path).toContain("/api/v1/");
        return;
      }
      if (node && typeof node === "object") {
        for (const [key, value] of Object.entries(node)) walk(value, `${path}.${key}`);
      }
    }
    walk(endpoints, "endpoints");
  });
});
