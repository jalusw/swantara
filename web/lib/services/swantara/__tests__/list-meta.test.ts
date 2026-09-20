import { describe, expect, it, vi } from "vitest";
import { Contacts } from "../identity";
import type { Contact, ListResult } from "../types";
import { withListMeta } from "../types";

describe("withListMeta", () => {
  it("preserves pagination meta alongside data", () => {
    const envelope = {
      success: true,
      message: "Retrieved.",
      data: { contacts: [{ id: 1 }] },
      meta: {
        timestamp: "2026-01-01T00:00:00Z",
        pagination: { page: 2, perPage: 10, total: 25, totalPages: 3 },
      },
    };

    expect(withListMeta(envelope)).toEqual({
      contacts: [{ id: 1 }],
      meta: envelope.meta,
    });
  });

  it("omits meta when the envelope has none", () => {
    const result = withListMeta({ success: true, message: "ok", data: { contacts: [] } });

    expect(result).toEqual({ contacts: [] });
    expect("meta" in result).toBe(false);
  });
});

describe("regression: list responses keep backend pagination meta", () => {
  it("should not drop meta from contacts list", async () => {
    const meta = {
      timestamp: "2026-01-01T00:00:00Z",
      pagination: { page: 2, perPage: 10, total: 25, totalPages: 3 },
    };
    const axios = {
      get: vi.fn(async () => ({ data: { data: { contacts: [] }, meta } })),
    };
    const service = new Contacts(axios as never);

    const result: ListResult<{ contacts: Contact[] }> = await service.list(1, {
      page: 2,
      size: 10,
    });

    expect(axios.get).toHaveBeenCalledWith(expect.any(String), {
      params: { page: 2, size: 10 },
    });
    expect(result.meta).toEqual(meta);
  });
});
