import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

import { getDiscordAvatar } from "@/lib/discord/getDiscordAvatar";

describe("getDiscordAvatar", () => {
  test.each([
    [undefined, "https://cdn.discordapp.com/embed/avatars/0.png"],
    [{ id: "1" }, "https://cdn.discordapp.com/embed/avatars/0.png"],
    [
      { id: "1", discriminator: "1237" },
      "https://cdn.discordapp.com/embed/avatars/2.png",
    ],
    [
      { id: "1", avatar: "abc" },
      "https://cdn.discordapp.com/avatars/1/abc.png?size=4096",
    ],
    [
      { id: "1", avatar: "a_abc" },
      "https://cdn.discordapp.com/avatars/1/a_abc.gif?size=4096",
    ],
  ])("%j → %s", (user, expected) => {
    expect(getDiscordAvatar(user)).toBe(expected);
  });
});

const member = { user: { id: "u1" }, roles: ["r1", "missing"] };
const guildRoles = [
  { id: "r1", name: "Mod", color: 0 },
  { id: "r9", name: "Other", color: 0 },
];

function respond(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("getDiscordMemberFromAPI", () => {
  beforeEach(() => {
    vi.resetModules();
  });
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  async function load(
    env: Record<string, string> = { GUILD_ID: "g1", BOT_TOKEN: "tok" },
  ) {
    for (const [k, v] of Object.entries(env)) vi.stubEnv(k, v);
    return (await import("@/lib/discord/getDiscordMemberFromAPI"))
      .getDiscordMemberFromAPI;
  }

  test("fetches the member and resolves their role objects", async () => {
    const fetchMock = vi.fn(async (url: string) =>
      url.endsWith("/roles") ? respond(200, guildRoles) : respond(200, member),
    );
    vi.stubGlobal("fetch", fetchMock);
    const getMember = await load();

    expect(await getMember("u1")).toEqual({
      ...member,
      enrichedRoles: [guildRoles[0]],
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "https://discord.com/api/guilds/g1/members/u1",
      { headers: { Authorization: "Bot tok" } },
    );
  });

  test("returns null without credentials, on API errors, and on network failures", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    expect(
      await (
        await load({ GUILD_ID: "", BOT_TOKEN: "" })
      )("u1"),
    ).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();

    vi.resetModules();
    fetchMock.mockImplementation(async () => respond(404, {}));
    expect(await (await load())("u1")).toBeNull();

    vi.resetModules();
    fetchMock.mockRejectedValue(new Error("offline"));
    expect(await (await load())("u1")).toBeNull();
  });
});

describe("getDiscordMember", () => {
  beforeEach(() => vi.resetModules());
  afterEach(() => vi.doUnmock("@/lib/discord/getDiscordMemberFromAPI"));

  test("uses the cache before asking Discord", async () => {
    const fromApi = vi.fn().mockResolvedValue({ ...member, enrichedRoles: [] });
    vi.doMock("@/lib/discord/getDiscordMemberFromAPI", () => ({
      getDiscordMemberFromAPI: fromApi,
    }));
    const { getDiscordMember } = await import("@/lib/discord/getDiscordMember");
    const { removeDiscordMemberFromCache } =
      await import("@/lib/discord/getDiscordMemberFromCache");

    const first = await getDiscordMember("u1");
    const second = await getDiscordMember("u1");

    expect(first.data).toEqual({ ...member, enrichedRoles: [] });
    expect(second.data).toEqual(first.data);
    expect(fromApi).toHaveBeenCalledTimes(1);
    expect(second.ttl).toBeGreaterThan(Date.now());

    removeDiscordMemberFromCache("u1");
    await getDiscordMember("u1");
    expect(fromApi).toHaveBeenCalledTimes(2);
  });

  test("reports unknown members without caching them", async () => {
    const fromApi = vi.fn().mockResolvedValue(null);
    vi.doMock("@/lib/discord/getDiscordMemberFromAPI", () => ({
      getDiscordMemberFromAPI: fromApi,
    }));
    const { getDiscordMember } = await import("@/lib/discord/getDiscordMember");

    expect(await getDiscordMember("nobody")).toEqual({ data: null, ttl: null });
    await getDiscordMember("nobody");
    expect(fromApi).toHaveBeenCalledTimes(2);
  });
});

test("ban checks are currently disabled and never block sign-in", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  const { isUserBanned } = await import("@/lib/discord/getDiscordBanStatus");

  expect(await isUserBanned("u1")).toBe(false);
  expect(fetchMock).not.toHaveBeenCalled();
  vi.unstubAllGlobals();
});
