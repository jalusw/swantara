import { describe, expect, it } from "vitest";
import { Auth, Contacts, Me, Users } from "../identity";

describe("identity service classes", () => {
  it("exports Auth with expected methods", () => {
    const proto = Auth.prototype;
    expect(typeof proto.login).toBe("function");
    expect(typeof proto.register).toBe("function");
    expect(typeof proto.refresh).toBe("function");
    expect(typeof proto.logout).toBe("function");
  });

  it("exports Me with expected methods", () => {
    const proto = Me.prototype;
    expect(typeof proto.me).toBe("function");
    expect(typeof proto.organizations).toBe("function");
    expect(typeof proto.permissions).toBe("function");
  });

  it("exports Contacts with expected methods", () => {
    const proto = Contacts.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.delete).toBe("function");
  });

  it("exports Users with expected methods", () => {
    const proto = Users.prototype;
    expect(typeof proto.list).toBe("function");
    expect(typeof proto.get).toBe("function");
    expect(typeof proto.create).toBe("function");
    expect(typeof proto.delete).toBe("function");
  });
});
