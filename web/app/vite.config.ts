import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
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
