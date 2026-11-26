import { afterEach, beforeEach, expect, test, vi } from "vitest";

beforeEach(() => vi.resetModules());
afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

test("logs a tagged JSON payload at the matching console level", async () => {
  const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
  const { logWarn } = await import("@/lib/logger/logger");

  logWarn("Auth:Denied", { userId: "u1" });

  const [tag, json] = warn.mock.calls[0]!;
  expect(tag).toBe("[Auth:Denied]");
  expect(JSON.parse(json as string)).toMatchObject({
    userId: "u1",
    timestamp: expect.any(String),
  });
});

test("logError includes the error message and extra fields", async () => {
  const error = vi.spyOn(console, "error").mockImplementation(() => {});
  const { logError } = await import("@/lib/logger/logger");

  logError("Thing:Broke", new Error("boom"), { path: "/x" });
  logError("Thing:Odd", "plain string");

  expect(JSON.parse(error.mock.calls[0]![1] as string)).toMatchObject({
    message: "boom",
    path: "/x",
    stack: expect.any(String),
  });
  expect(JSON.parse(error.mock.calls[1]![1] as string)).toMatchObject({
    message: "plain string",
    stack: null,
  });
});

test("sends a Discord webhook embed when webhook logging is enabled", async () => {
  vi.stubEnv("DEBUG_WEBHOOK", "true");
  vi.stubEnv("DISCORD_ERROR_WEBHOOK", "https://hooks.example/error");
  vi.spyOn(console, "error").mockImplementation(() => {});
  const fetchMock = vi
    .fn()
    .mockResolvedValue(new Response(null, { status: 204 }));
  vi.stubGlobal("fetch", fetchMock);
  const { logError } = await import("@/lib/logger/logger");

  logError("Thing:Broke", new Error("x".repeat(5000)));
  await vi.waitFor(() => expect(fetchMock).toHaveBeenCalled());

  const [url, init] = fetchMock.mock.calls[0]!;
  const embed = JSON.parse(init.body).embeds[0];
  expect(url).toBe("https://hooks.example/error");
  expect(embed.title).toBe("[Thing:Broke] [ERROR]");
  expect(embed.description.length).toBeLessThanOrEqual(4096);
  expect(embed.description.endsWith("...\n```")).toBe(true);
});

test("webhook failures fall back to the console", async () => {
  vi.stubEnv("DEBUG_WEBHOOK", "true");
  vi.stubEnv("DISCORD_INFO_WEBHOOK", "https://hooks.example/info");
  vi.spyOn(console, "info").mockImplementation(() => {});
  const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
  vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
  const { logInfo } = await import("@/lib/logger/logger");

  logInfo("Thing:Happened");
  await vi.waitFor(() =>
    expect(warn).toHaveBeenCalledWith(
      "webhook crashed, falling back to console:",
      expect.any(Error),
    ),
  );
});
