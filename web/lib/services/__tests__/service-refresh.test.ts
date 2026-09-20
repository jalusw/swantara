import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { resetSharedRefreshForTests, SwantaraService } from "@/lib/services/swantara";
import { server } from "@/lib/tests";

describe("regression: dashboard logout after login", () => {
  it("should not logout on concurrent 401s, single refresh then retry", async () => {
    resetSharedRefreshForTests();
    let refreshCalls = 0;
    let meCalls = 0;
    server.use(
      http.get("*/api/v1/me", ({ request }) => {
        meCalls += 1;
        if (meCalls <= 2) return new HttpResponse(null, { status: 401 });
        expect(request.headers.get("authorization")).toBe("Bearer new-access");
        return HttpResponse.json({
          success: true,
          message: "",
          data: { user: { id: 2, email: "jalu@mail.com" } },
        });
      }),
    );

    const service = new SwantaraService("http://swantara.test", {
      refreshToken: async () => {
        refreshCalls += 1;
        await new Promise((resolve) => setTimeout(resolve, 10));
        return { accessToken: "new-access", refreshToken: "new-refresh" };
      },
    });

    const [first, second] = await Promise.all([service.me.me(), service.me.me()]);

    expect(first.user.id).toBe(2);
    expect(second.user.id).toBe(2);
    expect(refreshCalls).toBe(1);
    vi.restoreAllMocks();
    resetSharedRefreshForTests();
  });
});
