// @vitest-environment happy-dom
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

const session = vi.fn();
const replace = vi.fn();
vi.mock("next-auth/react", () => ({
  useSession: () => session(),
  SessionProvider: ({ children }: { children: unknown }) => children,
}));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace, push: vi.fn() }),
}));

const { RoleContent } = await import("@auth/components/Gates/RoleContent");
const AuthGate = (await import("@auth/components/Gates/AuthGate")).default;

beforeEach(() => {
  session.mockReset();
  replace.mockReset();
});
afterEach(cleanup);

test("RoleContent shows children only to members with a matching role", () => {
  session.mockReturnValue({
    data: { user: { id: "u", roles: ["a"] } },
    status: "authenticated",
  });

  render(
    <>
      <p>
        <RoleContent roles={["a", "b"]}>any-match</RoleContent>
      </p>
      <p>
        <RoleContent roles={["a", "b"]} match="all" fallback="all-fallback">
          all-match
        </RoleContent>
      </p>
      <p>
        <RoleContent roles={["z"]} fallback="no-access">
          hidden
        </RoleContent>
      </p>
    </>,
  );

  expect(screen.queryByText("any-match")).not.toBeNull();
  expect(screen.queryByText("all-match")).toBeNull();
  expect(screen.queryByText("all-fallback")).not.toBeNull();
  expect(screen.queryByText("no-access")).not.toBeNull();
  expect(screen.queryByText("hidden")).toBeNull();
});

test("AuthGate renders the page for a signed-in member with roles", () => {
  session.mockReturnValue({
    data: { user: { id: "u", roles: ["a"] } },
    status: "authenticated",
  });

  render(<AuthGate>secret page</AuthGate>);

  expect(screen.queryByText("secret page")).not.toBeNull();
  expect(replace).not.toHaveBeenCalled();
});

test.each([
  [
    { data: null, status: "unauthenticated" },
    "/auth/login?error=Unauthenticated",
  ],
  [
    { data: { user: { id: "u", roles: [] } }, status: "authenticated" },
    "/auth/login?error=InvalidSession",
  ],
])("AuthGate sends invalid sessions to login", async (state, target) => {
  session.mockReturnValue(state);

  render(<AuthGate>secret page</AuthGate>);

  await waitFor(() => expect(replace).toHaveBeenCalledWith(target));
});

test("AuthGate waits while the session loads", () => {
  session.mockReturnValue({ data: null, status: "loading" });

  render(<AuthGate>secret page</AuthGate>);

  expect(screen.queryByText("secret page")).toBeNull();
  expect(screen.queryByText("Warming up your session...")).not.toBeNull();
  expect(replace).not.toHaveBeenCalled();
});
