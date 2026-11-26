import { describe, expect, test } from "vitest";
import {
  extractRoleIds,
  extractRoleNames,
  getAllowedCards,
  getRoleById,
  getRoleName,
  hasAllRoles,
  hasAnyRole,
  isDiscordRole,
} from "@auth/core/roleUtils";
import { Roles } from "@auth/core/roleConstants";

const mod = { id: "m", name: "Moderator", color: 1 };
const user = { id: "u", roles: ["a", mod] };

describe("role helpers", () => {
  test("isDiscordRole recognises role objects only", () => {
    expect(isDiscordRole(mod)).toBe(true);
    expect(isDiscordRole({ id: "x", name: "x" })).toBe(false);
    expect(isDiscordRole("m")).toBe(false);
    expect(isDiscordRole(null)).toBe(false);
  });

  test("extracts ids and names from mixed role lists", () => {
    expect(extractRoleIds(["a", mod])).toEqual(["a", "m"]);
    expect(extractRoleIds()).toEqual([]);
    expect(extractRoleNames(["a", mod])).toEqual(["a", "Moderator"]);
  });

  test("hasAnyRole and hasAllRoles", () => {
    expect(hasAnyRole(user, ["m", "z"])).toBe(true);
    expect(hasAnyRole(user, ["z"])).toBe(false);
    expect(hasAnyRole(null, ["a"])).toBe(false);
    expect(hasAllRoles(user, ["a", "m"])).toBe(true);
    expect(hasAllRoles(user, ["a", "z"])).toBe(false);
    expect(hasAllRoles(undefined, ["a"])).toBe(false);
  });

  test("getAllowedCards keeps open cards and cards with any matching role group", () => {
    const cards = [
      { href: "/open", title: "", desc: "" },
      { href: "/mod", title: "", desc: "", roles: [["z"], ["m"]] },
      { href: "/admin", title: "", desc: "", roles: [["z"]] },
    ];

    expect(getAllowedCards(user, cards).map((c) => c.href)).toEqual([
      "/open",
      "/mod",
    ]);
  });

  test("getRoleById finds string and object roles", () => {
    expect(getRoleById(user, "a")).toBe("a");
    expect(getRoleById(user, "m")).toBe(mod);
    expect(getRoleById(user, "z")).toBeNull();
    expect(getRoleById(null, "a")).toBeNull();
  });

  test("getRoleName names the configured roles", () => {
    expect(getRoleName(Roles.MOD)).toBe("Moderator");
    expect(getRoleName(Roles.EVENT)).toBe("Event Staff");
    expect(getRoleName("unknown")).toBeNull();
  });
});
