import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  build: {
    sourcemap: true,
    target: "es2023",
  },
  test: {
    environment: "node",
  },
});
