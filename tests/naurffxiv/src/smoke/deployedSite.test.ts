import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, test } from "vitest";
import {
  publicPrerenderedRoutes,
  type PrerenderManifest,
} from "./prerenderedRoutes";

const baseUrl = process.env["SMOKE_BASE_URL"];
const manifestPath =
  process.env["SMOKE_PRERENDER_MANIFEST"] ??
  fileURLToPath(
    new URL(
      "../../../../services/naurffxiv/.next/prerender-manifest.json",
      import.meta.url,
    ),
  );

function loadRoutes(): string[] {
  const manifest = JSON.parse(
    readFileSync(manifestPath, "utf8"),
  ) as PrerenderManifest;
  return publicPrerenderedRoutes(manifest);
}

const routes = baseUrl ? loadRoutes() : [];

describe.skipIf(!baseUrl)("deployed site", () => {
  test("the build prerendered content pages to check", () => {
    expect(routes.length).toBeGreaterThan(2);
  });

  test.each(routes)("serves %s", async (route) => {
    const response = await fetch(new URL(route, baseUrl));
    expect(response.status).toBe(200);
  });
});
