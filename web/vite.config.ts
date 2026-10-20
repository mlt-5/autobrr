import { fileURLToPath, URL } from "node:url";
import { defineConfig, loadEnv, ConfigEnv } from "vite";
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import tailwindcss from '@tailwindcss/vite'

interface PreRenderedAsset {
  name: string | undefined;
  source: string | Uint8Array;
  type: 'asset';
}

interface PreRenderedChunk {
  name: string;
  facadeModuleId: string | null;
  type: 'chunk';
}

// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv) => {
  // early load .env file
  // import.meta.env.VITE_NAME available here with: process.env.VITE_NAME
  process.env = { ...process.env, ...loadEnv(mode, process.cwd()) };

  return defineConfig({
    base: "/",
    plugins: [react(), svgr(), tailwindcss()],
    resolve: {
      alias: [
        { find: "@", replacement: fileURLToPath(new URL("./src/", import.meta.url)) },
        { find: "@app", replacement: fileURLToPath(new URL("./src/", import.meta.url)) },
        { find: "@components", replacement: fileURLToPath(new URL("./src/components", import.meta.url)) },
        { find: "@forms", replacement: fileURLToPath(new URL("./src/forms", import.meta.url)) },
        { find: "@hooks", replacement: fileURLToPath(new URL("./src/hooks", import.meta.url)) },
        { find: "@api", replacement: fileURLToPath(new URL("./src/api", import.meta.url)) },
        { find: "@screens", replacement: fileURLToPath(new URL("./src/screens", import.meta.url)) },
        { find: "@utils", replacement: fileURLToPath(new URL("./src/utils", import.meta.url)) },
        { find: "@types", replacement: fileURLToPath(new URL("./src/types", import.meta.url)) },
        { find: "@domain", replacement: fileURLToPath(new URL("./src/domain", import.meta.url)) }
      ]
    },
    build: {
      sourcemap: true,
      rollupOptions: {
        output: {
          assetFileNames: (chunkInfo: PreRenderedAsset) => {
            if (chunkInfo.name === "Inter-Variable.woff2") {
              return "assets/[name][extname]";
            }
            return "assets/[name]-[hash][extname]";
          },
          chunkFileNames: (chunkInfo: PreRenderedChunk) => {
            const locale = chunkInfo.facadeModuleId?.match(/\/i18n\/locales\/([^/]+)\/([^/]+)\.json$/);
            if (locale) {
              return `assets/locales/${locale[1]}-${locale[2]}-[hash].js`;
            }
            return "assets/[name]-[hash].js";
          }
        },
      }
    }
  });
};
