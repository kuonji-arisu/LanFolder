import { defineConfig } from "vite";
import type { Plugin } from "vite";
import vue from "@vitejs/plugin-vue";
import wails from "@wailsio/runtime/plugins/vite";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

function appIconAsset(): Plugin {
  const iconPath = resolve(__dirname, "../build/appicon.png");

  return {
    name: "lanfolder-app-icon",
    configureServer(server) {
      server.middlewares.use("/appicon.png", (_req, res, next) => {
        try {
          res.statusCode = 200;
          res.setHeader("Content-Type", "image/png");
          res.setHeader("Cache-Control", "no-cache");
          res.end(readFileSync(iconPath));
        } catch {
          next();
        }
      });
    },
    generateBundle() {
      this.emitFile({
        type: "asset",
        fileName: "appicon.png",
        source: readFileSync(iconPath),
      });
    },
  };
}

export default defineConfig({
  plugins: [appIconAsset(), vue(), wails("./bindings")],
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
  build: {
    rollupOptions: {
      input: {
        app: resolve(__dirname, "index.html"),
        web: resolve(__dirname, "web.html"),
      },
      onwarn(warning, defaultHandler) {
        if (warning.code === "INVALID_ANNOTATION" && warning.id?.includes("@vueuse/core")) {
          return;
        }
        defaultHandler(warning);
      },
    },
  },
  test: {
    environment: "happy-dom",
  },
});
