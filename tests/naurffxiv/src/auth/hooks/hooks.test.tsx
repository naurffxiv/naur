// @vitest-environment happy-dom
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, test, vi } from "vitest";

const session = vi.fn();
const signIn = vi.fn();
const push = vi.fn();
const replace = vi.fn();
let params = new URLSearchParams();
vi.mock("next-auth/react", () => ({
  useSession: () => session(),
  signIn: (...a: unknown[]) => signIn(...a),
}));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => params,
}));

const { useHasAllRoles, useHasAnyRole, useRoleBasedContent, useUserRoles } =
  await import("@auth/hooks/useRoles");
const { useAuthError } = await import("@auth/hooks/useAuthError");
const { useAutoLogin } = await import("@auth/hooks/useAutoLogin");
const { useSessionRedirect } = await import("@auth/hooks/useSessionRedirect");

const signedIn = (roles: unknown[]) => ({
  data: { user: { id: "u1", roles } },
  status: "authenticated",
});
const signedOut = { data: null, status: "unauthenticated" };

beforeEach(() => {
  session.mockReset();
  signIn.mockReset();
  push.mockReset();
  params = new URLSearchParams();
});

describe("role hooks", () => {
  test("check the signed-in user's roles", () => {
    session.mockReturnValue(signedIn(["a", { id: "b", name: "B", color: 0 }]));

    expect(renderHook(() => useHasAnyRole(["b", "z"])).result.current).toBe(
      true,
    );
    expect(renderHook(() => useHasAllRoles(["a", "z"])).result.current).toBe(
      false,
    );
    expect(renderHook(() => useUserRoles()).result.current).toEqual(["a", "b"]);
    const items = [
      { id: 1 },
      { id: 2, roles: [["b"]] },
      { id: 3, roles: [["z"]] },
    ];
    expect(
      renderHook(() => useRoleBasedContent(items)).result.current.map(
        (i) => i.id,
      ),
    ).toEqual([1, 2]);
  });

  test("grant nothing when signed out", () => {
    session.mockReturnValue(signedOut);

    expect(renderHook(() => useHasAnyRole(["a"])).result.current).toBe(false);
    expect(renderHook(() => useHasAllRoles([])).result.current).toBe(false);
    expect(renderHook(() => useUserRoles()).result.current).toEqual([]);
    expect(
      renderHook(() =>
        useRoleBasedContent<{ id: number; roles?: string[][] }>([{ id: 1 }]),
      ).result.current,
    ).toEqual([]);
  });
});

describe("useAuthError", () => {
  test.each([
    ["AccessDenied", "unauthenticated", "Access Denied"],
    ["Banned", "unauthenticated", "Banned"],
    ["Callback", "authenticated", "Oops"],
    ["Callback", "unauthenticated", "Error"],
  ])("heading for %s", (error, status, heading) => {
    params = new URLSearchParams({ error });
    session.mockReturnValue({ data: null, status });

    expect(renderHook(() => useAuthError()).result.current.heading).toBe(
      heading,
    );
  });

  test("prefers the error description, then the message, then a default", () => {
    session.mockReturnValue(signedOut);

    params = new URLSearchParams({
      error: "OAuthCallback: state: mismatch",
      message: "ignored",
    });
    expect(renderHook(() => useAuthError()).result.current).toMatchObject({
      errorCode: "OAuthCallback",
      displayMessage: "state: mismatch",
    });

    params = new URLSearchParams({ error: "Weird", message: "Try%20again" });
    expect(renderHook(() => useAuthError()).result.current).toMatchObject({
      displayMessage: "Try again",
      decodedMessage: "Try again",
    });

    params = new URLSearchParams({ message: "%E0%A4%A" });
    expect(renderHook(() => useAuthError()).result.current).toMatchObject({
      errorCode: "Unknown",
      decodedMessage: "%E0%A4%A",
    });
  });
});

describe("useAutoLogin", () => {
  test("starts Discord sign-in for signed-out visitors who asked for it", async () => {
    const { result } = renderHook(() =>
      useAutoLogin("unauthenticated", true, "/dashboard"),
    );

    await waitFor(() =>
      expect(signIn).toHaveBeenCalledWith("discord", {
        callbackUrl: "/dashboard",
      }),
    );
    expect(result.current).toBe(true);
  });

  test.each([
    ["authenticated", true, ""],
    ["unauthenticated", false, ""],
    ["unauthenticated", true, "SessionExpired"],
  ])(
    "does nothing when status=%s auto=%s error=%s",
    async (status, auto, error) => {
      params = new URLSearchParams(error ? { error } : {});

      const { result } = renderHook(() => useAutoLogin(status, auto, "/"));

      await act(async () => {});
      expect(signIn).not.toHaveBeenCalled();
      expect(result.current).toBe(false);
    },
  );

  test("allows retrying when sign-in fails", async () => {
    signIn.mockRejectedValue(new Error("popup blocked"));

    const { result } = renderHook(() =>
      useAutoLogin("unauthenticated", true, "/"),
    );

    await waitFor(() => expect(signIn).toHaveBeenCalled());
    await waitFor(() => expect(result.current).toBe(false));
  });
});

describe("useSessionRedirect", () => {
  test("sends signed-in staff on to where they were going", async () => {
    const { result } = renderHook(() =>
      useSessionRedirect(
        signedIn(["a"]).data as never,
        "authenticated",
        "/mod-portal",
      ),
    );

    await waitFor(() => expect(push).toHaveBeenCalledWith("/mod-portal"));
    expect(result.current).toBe(true);
  });

  test.each([
    ["loading", signedIn(["a"]).data],
    ["unauthenticated", null],
    ["authenticated", signedIn([]).data],
  ])("stays put when status=%s", async (status, data) => {
    renderHook(() => useSessionRedirect(data as never, status, "/"));

    await act(async () => {});
    expect(push).not.toHaveBeenCalled();
  });
});
