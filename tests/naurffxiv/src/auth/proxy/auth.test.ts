import { NextRequest } from "next/server";
import { beforeEach, describe, expect, test, vi } from "vitest";

const getToken = vi.fn();
vi.mock("next-auth/jwt", () => ({
  getToken: (...args: unknown[]) => getToken(...args),
}));

const { handleProtectedRoutesAuth } = await import("@auth/proxy/auth");
const { proxy } = await import("@/proxy");
const { Roles } = await import("@auth/core/roleConstants");

function request(path: string) {
  return new NextRequest(new URL(path, "https://naurffxiv.com"));
}

function location(response: Response | null) {
  const url = new URL(response!.headers.get("location")!);
  return `${url.pathname}?${url.searchParams.toString()}`;
}

beforeEach(() => {
  getToken.mockReset();
});

describe("handleProtectedRoutesAuth", () => {
  test.each([
    "/",
    "/ultimate/fru",
    "/auth/login",
    "/api/auth/callback/discord",
  ])("lets %s through without checking", async (path) => {
    expect(await handleProtectedRoutesAuth(request(path))).toBeNull();
    expect(getToken).not.toHaveBeenCalled();
  });

  test("sends signed-out visitors to the access denied page", async () => {
    getToken.mockResolvedValue(null);

    const response = await handleProtectedRoutesAuth(
      request("/mod-portal/notes"),
    );

    expect(location(response)).toBe(
      "/auth/error?from=%2Fmod-portal%2Fnotes&error=AccessDenied",
    );
  });

  test("lets members with an allowed role in and remembers the page", async () => {
    getToken.mockResolvedValue({
      id: "u",
      name: "mod",
      roles: [{ id: Roles.MOD, name: "Moderator", color: 0 }],
    });

    const response = await handleProtectedRoutesAuth(
      request("/mod-portal/notes"),
    );

    expect(response!.headers.get("x-middleware-next")).toBe("1");
    expect(response!.cookies.get("lastPath")?.value).toBe("/mod-portal/notes");
  });

  test("turns away members without an allowed role", async () => {
    getToken.mockResolvedValue({ id: "u", roles: [Roles.EVENT] });

    const response = await handleProtectedRoutesAuth(request("/admin"));

    expect(location(response)).toBe(
      "/auth/error?from=%2Fadmin&error=AccessDenied",
    );
  });
});

describe("proxy", () => {
  test("passes unprotected requests through", async () => {
    const response = await proxy(request("/"));

    expect(response.headers.get("x-middleware-next")).toBe("1");
  });

  test("redirects to login when the auth check crashes", async () => {
    getToken.mockImplementation(async () => {
      throw new Error("bad secret");
    });

    const response = await proxy(request("/admin"));

    expect(location(response)).toBe("/auth/login?error=UnexpectedProxyError");
    expect(response.headers.get("X-Proxy-Error")).toBe("true");
  });
});
