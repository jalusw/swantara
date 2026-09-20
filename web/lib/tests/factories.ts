export function buildLoginPayload(overrides: { email?: string; password?: string } = {}) {
  return { email: "demo@swantara.local", password: "secret", ...overrides };
}
