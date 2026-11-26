// @vitest-environment happy-dom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

const signIn = vi.fn();
const push = vi.fn();
vi.mock("next-auth/react", () => ({
  signIn: (...a: unknown[]) => signIn(...a),
}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));
vi.mock("next/image", () => ({
  default: ({ unoptimized, priority, ...props }: Record<string, unknown>) => (
    <img
      {...props}
      data-unoptimized={String(unoptimized)}
      data-priority={String(priority)}
    />
  ),
}));

const { LoginButton } = await import("@auth/components/Login/LoginButton");
const { LoginErrorMessage } =
  await import("@auth/components/Login/LoginErrorMessage");
const LogoutButton = (await import("@auth/components/Logout/LogoutButton"))
  .default;
const UserAvatar = (await import("@auth/components/Elements/UserAvatar"))
  .default;
const ReleaseCard = (await import("@/components/Changelog/ReleaseCard"))
  .default;

beforeEach(() => {
  signIn.mockReset().mockResolvedValue(undefined);
  push.mockReset();
});
afterEach(cleanup);

describe("login and logout", () => {
  test("LoginButton signs in with Discord and returns to the callback page", () => {
    render(<LoginButton callbackUrl="/dashboard" />);

    fireEvent.click(screen.getByText("Sign in with Discord"));

    expect(signIn).toHaveBeenCalledWith("discord", {
      callbackUrl: "/dashboard",
    });
  });

  test("LoginErrorMessage shows the error and a way back", () => {
    const { container, rerender } = render(
      <LoginErrorMessage message={null} fromPage="/x" />,
    );
    expect(container.textContent).toBe("");

    rerender(
      <LoginErrorMessage message="Access denied" fromPage="/mod-portal" />,
    );
    expect(screen.getByText("Access denied")).toBeTruthy();
    expect(
      screen.getByText("Return to previous page").getAttribute("href"),
    ).toBe("/mod-portal");

    rerender(<LoginErrorMessage message="Access denied" fromPage={null} />);
    expect(screen.queryByText("Return to previous page")).toBeNull();
  });

  test("LogoutButton goes to the logout page", () => {
    render(<LogoutButton />);

    fireEvent.click(screen.getByText("Logout"));

    expect(push).toHaveBeenCalledWith("/auth/logout");
  });
});

describe("UserAvatar", () => {
  test("shows the Discord avatar with a readable alt text", () => {
    render(
      <UserAvatar
        user={{ id: "1", name: " Alex ", avatar: "a_abc" } as never}
        size={32}
      />,
    );

    const img = screen.getByAltText("Alex's avatar");
    expect(img.getAttribute("src")).toBe(
      "https://cdn.discordapp.com/avatars/1/a_abc.gif?size=4096",
    );
    expect(img.getAttribute("width")).toBe("32");
    expect(img.getAttribute("data-unoptimized")).toBe("true");
  });

  test("falls back to a generic alt text", () => {
    render(<UserAvatar user={{ id: "1" } as never} />);

    expect(screen.getByAltText("User's avatar")).toBeTruthy();
  });
});

describe("ReleaseCard", () => {
  const release = {
    version: "v1.2.0",
    date: "2026-01-02T03:04:05Z",
    body: "## Added\n- dark mode",
  };

  test("older releases toggle when the header is clicked, but not via the version link", () => {
    const toggle = vi.fn();
    render(
      <ReleaseCard
        release={release}
        isExpanded={false}
        toggleRelease={toggle}
        isLatest={false}
      />,
    );

    fireEvent.click(screen.getByText("v1.2.0"));
    expect(toggle).not.toHaveBeenCalled();

    fireEvent.click(screen.getByText("v1.2.0").closest("div.cursor-pointer")!);
    expect(toggle).toHaveBeenCalledWith("v1.2.0");
    expect(screen.getByText("v1.2.0").getAttribute("href")).toBe(
      "https://github.com/naurffxiv/naurffxiv/releases/tag/v1.2.0",
    );
  });

  test("the latest release is marked and always open", () => {
    const toggle = vi.fn();
    render(
      <ReleaseCard
        release={release}
        isExpanded={false}
        toggleRelease={toggle}
        isLatest
      />,
    );

    expect(screen.getByText("Latest")).toBeTruthy();
    expect(screen.getByText("dark mode")).toBeTruthy();
    fireEvent.click(screen.getByText("Latest"));
    expect(toggle).not.toHaveBeenCalled();
  });

  test("handles releases without notes or versions", () => {
    const { container, rerender } = render(
      <ReleaseCard
        release={{ version: "v1.0.0", date: null, body: null }}
        isExpanded
        toggleRelease={vi.fn()}
        isLatest={false}
      />,
    );
    expect(screen.getByText("No release notes yet.")).toBeTruthy();

    rerender(
      <ReleaseCard
        release={{ version: "", date: null, body: null }}
        isExpanded
        toggleRelease={vi.fn()}
        isLatest={false}
      />,
    );
    expect(container.textContent).toBe("");
  });
});
