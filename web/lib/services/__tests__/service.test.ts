import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { SwantaraService, SwantaraUnauthorizedError } from "@/lib/services/swantara";
import { server } from "@/lib/tests";

const service = new SwantaraService("http://swantara.test");

describe("SwantaraService", () => {
  it("snake-cases request bodies", async () => {
    let receivedBody: Record<string, unknown> | undefined;

    server.use(
      http.post("*/api/v1/organizations", async ({ request }) => {
        receivedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(
          {
            success: true,
            message: "Organization created successfully.",
            data: { organization: { id: 1 } },
          },
          { status: 201 },
        );
      }),
    );

    await service.organizations.create({
      name: "Swantara",
      legalName: "Swantara Ltd",
    });

    expect(receivedBody).toEqual({
      name: "Swantara",
      legal_name: "Swantara Ltd",
    });
  });

  it("deep-camel-cases response data", async () => {
    server.use(
      http.get("*/api/v1/organizations/1", () =>
        HttpResponse.json({
          success: true,
          message: "Organization retrieved successfully.",
          data: {
            organization: {
              id: 1,
              legal_name: "Swantara Ltd",
              created_at: "2026-08-01T00:00:00Z",
            },
          },
        }),
      ),
    );

    const { organization } = await service.organizations.get(1);

    expect(organization).toMatchObject({
      legalName: "Swantara Ltd",
      createdAt: "2026-08-01T00:00:00Z",
    });
  });

  it("forwards FormData untouched on attachment upload", async () => {
    server.use(
      http.post("*/api/v1/organizations/1/attachments", async ({ request }) => {
        expect(request.headers.get("content-type")).toContain("multipart/form-data");
        const body = await request.text();
        expect(body).toContain('name="file"');
        return HttpResponse.json(
          {
            success: true,
            message: "Attachment uploaded successfully.",
            data: {
              attachment: {
                id: 1,
                filename: "avatar.png",
                mime_type: "image/png",
                byte_size: 1024,
              },
            },
          },
          { status: 201 },
        );
      }),
    );

    const { attachment } = await service.attachments.upload(1, {
      ownerType: "user",
      ownerId: 1,
      file: new File(["x"], "avatar.png"),
    });

    expect(attachment).toMatchObject({
      filename: "avatar.png",
      mimeType: "image/png",
      byteSize: 1024,
    });
  });

  it("forwards list query parameters", async () => {
    server.use(
      http.get("*/api/v1/users", ({ request }) => {
        expect(request.url).toContain("page=2");
        expect(request.url).toContain("size=50");
        return HttpResponse.json({
          success: true,
          message: "Users data retrieved successfully.",
          data: { users: [] },
        });
      }),
    );

    const result = await service.users.list({ page: 2, size: 50 });

    expect(result).toEqual({ users: [] });
  });

  it("maps error responses to Swantara errors", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json(
          {
            success: false,
            message: "Invalid credentials provided",
            error_code: "ERR_UNAUTHORIZED",
            error: "invalid credentials",
          },
          { status: 401 },
        ),
      ),
    );

    const error = await service.auth
      .login({ email: "a@b.com", password: "secret" })
      .catch((e: unknown) => e);

    expect(error).toBeInstanceOf(SwantaraUnauthorizedError);
    expect(error).toMatchObject({
      message: "Invalid credentials provided",
      code: "ERR_UNAUTHORIZED",
      status: 401,
    });
  });

  it("resolves no-content responses", async () => {
    server.use(http.delete("*/api/v1/users/1", () => new HttpResponse(null, { status: 204 })));

    await expect(service.users.delete(1)).resolves.toBeUndefined();
  });
});
