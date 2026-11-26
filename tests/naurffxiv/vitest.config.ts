import { defineConfig } from "vitest/config";
import path from "path";

const service = path.resolve(import.meta.dirname, "../../services/naurffxiv");

export default defineConfig({
  test: {
    environment: "node",
    globals: true,
    setupFiles: ["./vitest.setup.ts"],
    restoreMocks: true,
  },
  resolve: {
    alias: {
      "@auth": path.resolve(service, "src/auth"),
      "@naur/shared": path.resolve(
        import.meta.dirname,
        "../../packages/shared",
      ),
      "@": path.resolve(service, "src"),
    },
    // Components under test and the tests must share one React instance.
    dedupe: ["react", "react-dom", "next-auth"],
  },
});
