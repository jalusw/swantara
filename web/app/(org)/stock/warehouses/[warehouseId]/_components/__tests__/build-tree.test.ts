import { describe, expect, it } from "vitest";

type LocationRow = {
  id: string;
  name: string;
  code: string | null;
  parentId: string | null;
  usage: string;
  warehouseId: string | null;
};

type TreeNode = {
  id: string;
  label: string;
  children: TreeNode[];
};

function buildTree(locations: LocationRow[]): TreeNode[] {
  const childrenMap = new Map<string | null, LocationRow[]>();

  for (const location of locations) {
    const key = location.parentId;
    if (!childrenMap.has(key)) {
      childrenMap.set(key, []);
    }
    childrenMap.get(key)?.push(location);
  }

  function buildNodes(parentId: string | null): TreeNode[] {
    return (childrenMap.get(parentId) ?? []).map((location) => ({
      id: location.id,
      label: location.name,
      children: buildNodes(location.id),
    }));
  }

  return buildNodes(null);
}

describe("buildTree", () => {
  it("returns an empty array when no locations are provided", () => {
    expect(buildTree([])).toEqual([]);
  });

  it("builds a flat tree from root-level locations", () => {
    const locations: LocationRow[] = [
      {
        id: "1",
        name: "WH-1",
        code: "W1",
        parentId: null,
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "2",
        name: "WH-2",
        code: "W2",
        parentId: null,
        usage: "inventory",
        warehouseId: "10",
      },
    ];
    const tree = buildTree(locations);
    expect(tree).toHaveLength(2);
    expect(tree[0]!.id).toBe("1");
    expect(tree[0]!.children).toEqual([]);
    expect(tree[1]!.id).toBe("2");
  });

  it("nests child locations under their parent", () => {
    const locations: LocationRow[] = [
      {
        id: "1",
        name: "Root",
        code: null,
        parentId: null,
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "2",
        name: "Shelf A",
        code: "SA",
        parentId: "1",
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "3",
        name: "Shelf B",
        code: "SB",
        parentId: "1",
        usage: "inventory",
        warehouseId: "10",
      },
    ];
    const tree = buildTree(locations);
    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe("1");
    expect(tree[0]!.children).toHaveLength(2);
    expect(tree[0]!.children[0]!.id).toBe("2");
    expect(tree[0]!.children[1]!.id).toBe("3");
  });

  it("supports multi-level nesting", () => {
    const locations: LocationRow[] = [
      {
        id: "1",
        name: "Root",
        code: null,
        parentId: null,
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "2",
        name: "Level 1",
        code: null,
        parentId: "1",
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "3",
        name: "Level 2",
        code: null,
        parentId: "2",
        usage: "inventory",
        warehouseId: "10",
      },
    ];
    const tree = buildTree(locations);
    expect(tree[0]!.children[0]!.children[0]!.id).toBe("3");
  });

  it("filters locations by warehouseId before building", () => {
    const locations: LocationRow[] = [
      {
        id: "1",
        name: "WH1 Loc",
        code: null,
        parentId: null,
        usage: "inventory",
        warehouseId: "10",
      },
      {
        id: "2",
        name: "WH2 Loc",
        code: null,
        parentId: null,
        usage: "inventory",
        warehouseId: "20",
      },
    ];
    const filtered = locations.filter((l) => l.warehouseId === "10");
    const tree = buildTree(filtered);
    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe("1");
  });
});
