import { expect, test } from "vitest";
import {
  getAllowedRolesForPath,
  routeRoleAccessMap,
} from "@auth/core/routeRoles";
import { Roles } from "@auth/core/roleConstants";
import { protectedRoutes } from "@auth/proxy/routes.config";
import { config } from "@/proxy";

test("protected routes require their configured roles, including nested pages", () => {
  expect(getAllowedRolesForPath("/admin")).toEqual([Roles.ADMIN]);
  expect(getAllowedRolesForPath("/mod-portal/notes")).toEqual([
    Roles.ADMIN,
    Roles.MOD,
  ]);
  expect(getAllowedRolesForPath("/dashboard")).toEqual([
    Roles.ADMIN,
    Roles.MOD,
    Roles.DEV,
    Roles.EVENT,
  ]);
});

test("unprotected routes are open", () => {
  expect(getAllowedRolesForPath("/")).toEqual([]);
  expect(getAllowedRolesForPath("/ultimate/fru")).toEqual([]);
  expect(getAllowedRolesForPath()).toEqual([]);
});

test("route role map, protected route list and proxy matcher agree", () => {
  const roleMapped = Object.keys(routeRoleAccessMap).sort();

  expect([...protectedRoutes].sort()).toEqual(roleMapped);
  expect(config.matcher.map((m) => m.replace("/:path*", "")).sort()).toEqual(
    roleMapped,
  );
});
