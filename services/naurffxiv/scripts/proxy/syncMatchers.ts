import chalk from "chalk";
import fs from "fs";
import path from "path";

const ROUTES_FILE = path.resolve("src/auth/proxy/routes.config.ts");
const PROXY_FILE = path.resolve("src/proxy.ts");

function extractRoutes(): string[] {
  const content = fs.readFileSync(ROUTES_FILE, "utf8");

  const regexes = [
    // Matches: export const protectedRoutes: readonly string[] = [ ... ]
    // Captures everything between the opening and closing brackets
    // \s+ = one or more whitespace characters
    // \s* = zero or more whitespace characters
    // [\s\S]*? = any character (including newlines), non-greedy match
    // ([\s\S]*?) = capture group for array contents
    /export\s+const\s+protectedRoutes\s*:\s*readonly\s*string\[\]\s*=\s*\[([\s\S]*?)\]/,

    // Matches: export const protectedRoutes = [ ... ] as const
    // Alternative syntax for the same protectedRoutes array
    /export\s+const\s+protectedRoutes\s*=\s*\[([\s\S]*?)\]\s*as\s+const/,
  ];

  let match: RegExpExecArray | null = null;
  for (const regex of regexes) {
    match = regex.exec(content);
    if (match) break;
  }

  if (!match) {
    console.error("Could not parse protectedRoutes array.");
    console.log("First 300 char of content:\n", content.slice(0, 300));
    throw new Error("Could not parse protectedRoutes array");
  }

  // Extract array contents and clean up
  return match[1]
    .split(",") // Split by comma to get individual routes
    .map((s) => s.trim()) // Remove leading/trailing whitespace
    .map((s) => s.replace(/^['"`]|['"`]$/g, "")) // Remove quotes from start/end
    .filter((s) => s.length > 0 && !s.startsWith("//")); // Remove empty strings and comments
}

function generateUpdatedMatchers(
  existing: string[],
  fromRoutes: string[],
): string[] {
  const seen = new Set(existing);
  // Add /:path* suffix to each route (Next.js proxy matcher syntax)
  for (const route of fromRoutes.map((r) => `${r}/:path*`)) {
    if (!seen.has(route)) {
      existing.push(route);
      seen.add(route);
    }
  }
  return existing;
}

function syncMatchers(): void {
  const routes = extractRoutes();

  const proxyContent = fs.readFileSync(PROXY_FILE, "utf8");

  // Matches: matcher: [ ... ]
  // Captures everything between the opening and closing brackets of the matcher array
  // [\s\S]*? = any character (including newlines), non-greedy match to find the shortest match
  const matcherRegex = /matcher:\s*\[([\s\S]*?)\]/;
  const match = matcherRegex.exec(proxyContent);

  if (!match) {
    console.error("Could not find matcher array in proxy.ts");
    process.exit(1);
  }

  // Extract existing matchers and clean up
  const existingMatchers = match[1]
    .split(",") // Split by comma to get individual matchers
    .map((s) => s.trim()) // Remove leading/trailing whitespace
    .map((s) => s.replace(/^['"`]|['"`]$/g, "")) // Remove quotes from start/end
    .filter((s) => s.length > 0 && !s.startsWith("//")); // Remove empty strings and comments

  // Merge existing matchers with new routes (avoiding duplicates)
  const combinedMatchers = generateUpdatedMatchers(existingMatchers, routes);

  // Match the file's existing line ending to avoid mixed LF/CRLF output
  const eol = proxyContent.includes("\r\n") ? "\r\n" : "\n";

  // Reconstruct the matcher array with proper formatting
  const newBlock = `matcher: [${eol}${combinedMatchers
    // Format each route with indentation, quotes, and trailing comma
    .map((r) => `    "${r}",`)
    // Join routes with newlines for readability
    .join(eol)}${eol}  ]`;

  // Replace the old matcher array with the new one in proxy.ts
  const updatedProxy = proxyContent.replace(matcherRegex, newBlock);

  fs.writeFileSync(PROXY_FILE, updatedProxy, "utf8");
  if (process.env.NODE_ENV === "development") {
    console.log(
      chalk.green(" Matchers appended into proxy.ts without duplicates."),
    );
  }
}

syncMatchers();
