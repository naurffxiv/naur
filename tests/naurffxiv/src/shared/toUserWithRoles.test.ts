import { expect, test } from "vitest";
import { AuthErrorType, toUserWithRoles } from "@naur/shared/types";

const role = {
  id: "r2",
  name: "Mod",
  color: 0,
  position: 1,
  permissions: "0",
  managed: false,
  mentionable: false,
  hoist: false,
};

test("toUserWithRoles merges raw role ids and enriched roles", () => {
  const member = {
    user: { id: "u1" },
    roles: ["r1", "r2"],
    enrichedRoles: [role],
  };

  expect(toUserWithRoles(member as never)).toEqual({
    id: "u1",
    roles: ["r1", "r2", role],
  });
});

test("toUserWithRoles handles missing members and role lists", () => {
  expect(toUserWithRoles(null)).toBeNull();
  expect(toUserWithRoles(undefined)).toBeNull();
  expect(toUserWithRoles({ user: { id: "u1" } } as never)).toEqual({
    id: "u1",
    roles: [],
  });
});

test("auth error codes map to themselves", () => {
  for (const [key, value] of Object.entries(AuthErrorType)) {
    expect(value).toBe(key);
  }
});
