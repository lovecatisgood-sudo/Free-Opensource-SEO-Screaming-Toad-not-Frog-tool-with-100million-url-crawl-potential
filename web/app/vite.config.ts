import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: { "/api": { target: "http://127.0.0.1:7332", changeOrigin: true, headers: { Origin: "http://127.0.0.1:7332" } } },
  },
  build: {
    outDir: "../../internal/webui/assets",
    emptyOutDir: true,
    sourcemap: false,
    target: "es2023",
  },
  test: {
    environment: "node",
  },
});
