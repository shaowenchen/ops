import { fileURLToPath, URL } from "url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  build: {
    outDir: "./dist",
    assetsDir: "assets",
  },
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        // target: "http://localhost:80",
        target: "https://ops-server.wps.cn",
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
