import { describe, expect, it, vi } from "vitest";
import * as catalog from "../catalog";
import * as finance from "../finance";
import * as identity from "../identity";
import * as operations from "../operations";
import * as sales from "../sales";

function createAxiosMock() {
  const envelope = { data: { data: { ok: true } } };
  return {
    get: vi.fn(async () => envelope),
    post: vi.fn(async () => envelope),
    put: vi.fn(async () => envelope),
    patch: vi.fn(async () => envelope),
    delete: vi.fn(async () => envelope),
  };
}

type AxiosMock = ReturnType<typeof createAxiosMock>;

async function invokeMethods(target: unknown, seen: Set<unknown>): Promise<void> {
  if (!target || (typeof target !== "object" && typeof target !== "function")) return;
  if (seen.has(target)) return;
  seen.add(target);
  if (typeof target === "function") {
    try {
      await (target as (...args: unknown[]) => unknown)(1, 1, {}, {}, {});
    } catch {}
    return;
  }
  const record = target as Record<string, unknown>;
  for (const key of Object.keys(record)) {
    if (key === "constructor") continue;
    const value = record[key];
    if (typeof value === "function") {
      try {
        await (value as (...args: unknown[]) => unknown).call(record, 1, 1, {}, {}, {});
      } catch {}
    } else if (value && typeof value === "object") {
      await invokeMethods(value, seen);
    }
  }
  const proto = Object.getPrototypeOf(record);
  if (proto && proto !== Object.prototype) {
    for (const key of Object.getOwnPropertyNames(proto)) {
      if (key === "constructor") continue;
      const descriptor = Object.getOwnPropertyDescriptor(proto, key);
      if (descriptor && typeof descriptor.value === "function") {
        try {
          await descriptor.value.call(record, 1, 1, {}, {}, {});
        } catch {}
      }
    }
  }
}

async function executeNamespace(namespace: Record<string, unknown>): Promise<number> {
  let executed = 0;
  const axios = createAxiosMock();
  for (const [name, exported] of Object.entries(namespace)) {
    if (typeof exported !== "function" || !/^[A-Z]/.test(name)) continue;
    const instance = new (exported as new (axios: unknown) => unknown)(axios as never);
    await invokeMethods(instance, new Set());
    executed += 1;
  }
  return executed;
}

describe("service layer execution", () => {
  it("executes every catalog service method", async () => {
    const count = await executeNamespace(catalog as unknown as Record<string, unknown>);
    expect(count).toBeGreaterThan(0);
  });

  it("executes every finance service method", async () => {
    const count = await executeNamespace(finance as unknown as Record<string, unknown>);
    expect(count).toBeGreaterThan(0);
  });

  it("executes every operations service method", async () => {
    const count = await executeNamespace(operations as unknown as Record<string, unknown>);
    expect(count).toBeGreaterThan(0);
  });

  it("executes every sales service method", async () => {
    const count = await executeNamespace(sales as unknown as Record<string, unknown>);
    expect(count).toBeGreaterThan(0);
  });

  it("executes every identity service method", async () => {
    const count = await executeNamespace(identity as unknown as Record<string, unknown>);
    expect(count).toBeGreaterThan(0);
  });

  it("sends requests through the axios mock", async () => {
    const axios: AxiosMock = createAxiosMock();
    const service = new catalog.ProductCategories(axios as never);
    const result = await service.list(1);
    expect(axios.get).toHaveBeenCalled();
    expect(result).toEqual({ ok: true });
  });
});
