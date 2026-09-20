import { expect, type Page } from "@playwright/test";

const TOKEN = "e2e-flow-token";
const ORG_ID = "1";

export async function seedOrg(page: Page) {
  await page.context().addCookies([
    {
      name: "access_token",
      value: TOKEN,
      domain: "localhost",
      path: "/",
    },
    {
      name: "swantara-active-org",
      value: ORG_ID,
      domain: "localhost",
      path: "/",
    },
  ]);
}

export async function navigateToOrg(page: Page, path: string) {
  await page.goto(path);
  await expect(page.getByRole("main").first()).toBeVisible();
}

export async function createContact(page: Page, name: string, role: "customer" | "supplier") {
  await navigateToOrg(page, "/contacts");
  await page
    .getByRole("button", { name: /create|add/i })
    .first()
    .click();
  await page.getByLabel(/name/i).fill(name);
  await page.getByRole("switch", { name: new RegExp(role, "i") }).click();
  await page
    .getByRole("button", { name: /save|create/i })
    .first()
    .click();
  await expect(page.getByText(name)).toBeVisible();
}

export async function createProduct(page: Page, name: string, price: number) {
  await navigateToOrg(page, "/products");
  await page
    .getByRole("button", { name: /create|add/i })
    .first()
    .click();
  await page.getByLabel(/name/i).fill(name);
  await page
    .getByLabel(/price|list price/i)
    .first()
    .fill(String(price));
  await page
    .getByRole("button", { name: /save|create/i })
    .first()
    .click();
  await expect(page.getByText(name)).toBeVisible();
}

export async function waitForStableUrl(page: Page, pattern: RegExp) {
  await page.waitForURL(pattern, { timeout: 10_000 });
}
