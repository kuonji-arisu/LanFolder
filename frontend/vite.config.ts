import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import wails from "@wailsio/runtime/plugins/vite";
import { resolve } from "node:path";

export default defineConfig({
  plugins: [vue(), wails("./bindings")],
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
