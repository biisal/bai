import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: [
        {
          find: /^shiki\/langs$/,
          replacement: fileURLToPath(
            new URL("./src/lib/shiki-langs.ts", import.meta.url)
          ),
        },
        {
          find: "@",
          replacement: fileURLToPath(new URL("./src", import.meta.url)),
        },
      ],
    },
    server: {
      port: 5173,
      proxy: {
        "/api": {
          target: env.SERVER_URL || "http://127.0.0.1:18082",
          changeOrigin: true,
        },
      },
    },
  };
});
