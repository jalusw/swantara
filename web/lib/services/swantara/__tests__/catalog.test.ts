import { describe, expect, it } from "vitest";
import { Inventory, PriceBooks, ProductCategories, Products } from "../catalog";

describe("catalog service classes", () => {
  it("exports ProductCategories with expected methods", () => {
    const proto = ProductCategories.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.update).toBe("function");
    expect(typeof proto.delete).toBe("function");
  });

  it("exports Products with expected methods", () => {
    const proto = Products.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.update).toBe("function");
    expect(typeof proto.delete).toBe("function");
  });

  it("exports PriceBooks with expected methods", () => {
    const proto = PriceBooks.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.update).toBe("function");
    expect(typeof proto.delete).toBe("function");
    expect(typeof proto.resolve).toBe("function");
  });

  it("exports Inventory with expected methods", () => {
    const proto = Inventory.prototype;
    expect(typeof proto.warehouses).toBe("function");
    expect(typeof proto.stockLocations).toBe("function");
    expect(typeof proto.stockMovements).toBe("function");
    expect(typeof proto.onHand).toBe("function");
    expect(typeof proto.availableToPromise).toBe("function");
  });
});
