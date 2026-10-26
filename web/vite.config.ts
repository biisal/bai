import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const devToken = process.env.VITE_BAI_TOKEN;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.VITE_BAI_ADDR ?? "http://127.0.0.1:18082",
        changeOrigin: true,
        headers: devToken ? { Authorization: `Bearer ${devToken}` } : undefined,
      },
    },
  },
});
