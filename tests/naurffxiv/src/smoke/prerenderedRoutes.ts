export interface PrerenderManifest {
  routes: Record<string, { srcRoute: string | null }>;
}

const CONTENT_SRC_ROUTE = "/[difficulty]/[[...slug]]";
const PUBLIC_STATIC_ROUTES = ["/", "/changelog"];

export function publicPrerenderedRoutes(manifest: PrerenderManifest): string[] {
  return Object.entries(manifest.routes)
    .filter(
      ([route, entry]) =>
        entry.srcRoute === CONTENT_SRC_ROUTE ||
        PUBLIC_STATIC_ROUTES.includes(route),
    )
    .map(([route]) => route)
    .sort();
}
