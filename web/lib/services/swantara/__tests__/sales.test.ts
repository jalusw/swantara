import { describe, expect, it } from "vitest";
import { CrmLeads, GiftCards, PosSessions, SaleOrders } from "../sales";

describe("sales service classes", () => {
  it("exports SaleOrders with expected methods", () => {
    const proto = SaleOrders.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.confirm).toBe("function");
    expect(typeof proto.invoice).toBe("function");
  });

  it("exports CrmLeads with expected methods", () => {
    const proto = CrmLeads.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.promote).toBe("function");
  });

  it("exports PosSessions with expected methods", () => {
    const proto = PosSessions.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.close).toBe("function");
  });

  it("exports GiftCards with expected methods", () => {
    const proto = GiftCards.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.redeem).toBe("function");
  });
});
