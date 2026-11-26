import type { JWT } from "next-auth/jwt";
import { beforeEach, describe, expect, test, vi } from "vitest";

const getDiscordMember = vi.fn();
const isUserBanned = vi.fn();
vi.mock("@/lib/discord/getDiscordMember", () => ({
  getDiscordMember: (id: string) => getDiscordMember(id),
}));
vi.mock("@/lib/discord/getDiscordBanStatus", () => ({
  isUserBanned: (id: string) => isUserBanned(id),
}));
vi.mock("@/lib/logger/logger", () => ({
  logDebug: vi.fn(),
  logError: vi.fn(),
  logWarn: vi.fn(),
}));

const { jwt, session, signIn } = await import("@auth/core/callbacks");
const { Roles } = await import("@auth/core/roleConstants");

const profile = {
  id: "u1",
  username: "alex",
  avatar: "abc",
  discriminator: "0",
};
const modRole = { id: Roles.MOD, name: "Moderator", color: 0 };

function member(roles: string[], enrichedRoles = [] as unknown[]) {
  return { data: { user: { id: "u1" }, roles, enrichedRoles }, ttl: 1234 };
}

function signInWithProfile() {
  return signIn({
    user: {} as never,
    account: null,
    profile: profile as never,
  });
}

beforeEach(() => {
  getDiscordMember.mockReset();
  isUserBanned.mockReset().mockResolvedValue(false);
});

describe("signIn", () => {
  test("allows members with a staff role", async () => {
    getDiscordMember.mockResolvedValue(member([Roles.MOD]));

    expect(await signInWithProfile()).toBe(true);
  });

  test("rejects members without a staff role, unknown members and missing profiles", async () => {
    getDiscordMember.mockResolvedValue(member(["someone"]));
    expect(await signInWithProfile()).toBe(false);

    getDiscordMember.mockResolvedValue({ data: null, ttl: null });
    expect(await signInWithProfile()).toBe(false);

    expect(await signIn({ user: {} as never, account: null })).toBe(false);
  });

  test("sends banned users to the banned error page", async () => {
    isUserBanned.mockResolvedValue(true);

    expect(await signInWithProfile()).toBe("/auth/error?error=Banned");
    expect(getDiscordMember).not.toHaveBeenCalled();
  });
});

describe("jwt", () => {
  test("copies the Discord profile and roles into the token", async () => {
    getDiscordMember.mockResolvedValue(member([Roles.MOD], [modRole]));

    const token = await jwt({ token: {} as JWT, profile: profile as never });

    expect(token).toEqual({
      id: "u1",
      name: "alex",
      avatar: "abc",
      discriminator: "0",
      roles: [modRole],
      cacheExpires: 1234,
      hasAccess: true,
    });
  });

  test("denies access when Discord lookup fails", async () => {
    getDiscordMember.mockRejectedValue(new Error("discord down"));

    const token = await jwt({ token: {} as JWT, profile: profile as never });

    expect(token).toMatchObject({
      id: "u1",
      roles: [],
      cacheExpires: null,
      hasAccess: false,
    });
  });

  test("leaves the token alone on refresh without a profile", async () => {
    expect(await jwt({ token: { id: "u1", hasAccess: true } })).toEqual({
      id: "u1",
      hasAccess: true,
    });
    expect(getDiscordMember).not.toHaveBeenCalled();
  });
});

test("session exposes the token's user fields", async () => {
  const result = await session({
    session: { expires: "soon" } as never,
    token: {
      id: "u1",
      name: "alex",
      avatar: "abc",
      discriminator: "0",
      roles: [modRole],
      cacheExpires: 1,
      hasAccess: true,
    },
  });

  expect(result.user).toEqual({
    id: "u1",
    name: "alex",
    avatar: "abc",
    discriminator: "0",
    roles: [modRole],
    cacheExpires: 1,
    hasAccess: true,
  });
});
