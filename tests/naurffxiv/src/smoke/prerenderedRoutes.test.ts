import { expect, test } from "vitest";
import {
  publicPrerenderedRoutes,
  type PrerenderManifest,
} from "./prerenderedRoutes";

const manifest: PrerenderManifest = {
  routes: {
    "/": { srcRoute: "/" },
    "/changelog": { srcRoute: "/changelog" },
    "/ultimate/fru/guide": { srcRoute: "/[difficulty]/[[...slug]]" },
    "/ultimate/kefka": { srcRoute: "/[difficulty]/[[...slug]]" },
    "/dashboard": { srcRoute: "/dashboard" },
    "/mod-portal/notes": { srcRoute: "/mod-portal/notes" },
    "/auth/login": { srcRoute: "/auth/login" },
    "/_not-found": { srcRoute: "/_not-found" },
    "/icon.png": { srcRoute: "/icon.png" },
  },
};

test("publicPrerenderedRoutes keeps content pages, the home page and the changelog", () => {
  expect(publicPrerenderedRoutes(manifest)).toEqual([
    "/",
    "/changelog",
    "/ultimate/fru/guide",
    "/ultimate/kefka",
  ]);
});

test("publicPrerenderedRoutes leaves out protected, auth and non-page routes", () => {
  const routes = publicPrerenderedRoutes(manifest);
  for (const route of [
    "/dashboard",
    "/mod-portal/notes",
    "/auth/login",
    "/_not-found",
    "/icon.png",
  ]) {
    expect(routes).not.toContain(route);
  }
});

test("publicPrerenderedRoutes returns nothing for an empty manifest", () => {
  expect(publicPrerenderedRoutes({ routes: {} })).toEqual([]);
});
