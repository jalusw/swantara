import { describe, expect, it } from "vitest";
import {
  ME_ORGANIZATIONS_QUERY_KEY,
  ME_QUERY_KEY,
  meQueryOptions,
  permissionsQueryKey,
} from "../me";

const fakeService = {
  me: {
    me: async () => ({ data: {} }),
    organizations: async () => ({ data: [] }),
    permissions: async () => ({ data: [] }),
  },
} as never;

describe("me query constants", () => {
  it("exports ME_QUERY_KEY", () => {
    expect(ME_QUERY_KEY).toEqual(["me"]);
  });

  it("exports ME_ORGANIZATIONS_QUERY_KEY", () => {
    expect(ME_ORGANIZATIONS_QUERY_KEY).toEqual(["me", "organizations"]);
  });
});

describe("me query options", () => {
  it("meQueryOptions returns valid query options", () => {
    const opts = meQueryOptions(fakeService);
    expect(opts.queryKey).toEqual(ME_QUERY_KEY);
    expect(typeof opts.queryFn).toBe("function");
  });

  it("permissionsQueryKey includes organization id", () => {
    expect(permissionsQueryKey(42)).toEqual(["permissions", 42, {}]);
  });
});
