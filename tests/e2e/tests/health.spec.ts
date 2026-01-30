import { test, expect } from "@playwright/test";

test("basic health check", async ({ page }) => {
  // Navigate to the frontend
  await page.goto("/");

  // Verify API health endpoint
  const response = await page.request.get("http://localhost:3000/api/health");
  expect(response.status()).toBe(200);
});
