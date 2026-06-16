import { defineConfig } from "vite";

export default defineConfig({
  base: "/",
  server: {
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8090",
      "/health": "http://127.0.0.1:8090",
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
