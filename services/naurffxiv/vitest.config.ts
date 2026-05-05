import path from "path";
import react from "@vitejs/plugin-react";
import { defineConfig, type Plugin } from "vitest/config";

function mockStaticAssets(): Plugin {
  const assetRe = /\.(png|jpg|jpeg|gif|svg|webp|ico)$/;
  return {
    name: "mock-static-assets",
    resolveId(id) {
      if (assetRe.test(id)) return "\0mock-asset";
    },
    load(id) {
      if (id === "\0mock-asset") return 'export default ""';
    },
  };
}

export default defineConfig({ plugins: [mockStaticAssets(), react()] });
