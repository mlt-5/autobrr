import { fileURLToPath, URL } from "node:url";
import { defineConfig, loadEnv, ConfigEnv } from "vite";
import { VitePWA } from "vite-plugin-pwa";
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
    plugins: [react(), svgr(), tailwindcss(), VitePWA({
      injectRegister: null,
      selfDestroying: true,
      scope: "/",
      // strategies: "injectManifest",
      useCredentials: true,
      includeAssets: [
        // looks inside "public" folder 
        // manifest's icons are automatic added
        "favicon.ico"
      ],
      manifest: {
        name: "autobrr",
        short_name: "autobrr",
        description: "Automation for downloads.",
        theme_color: "#141415",
        background_color: "#141415",
        icons: [
          {
            src: "logo192.png",
            sizes: "192x192",
            type: "image/png"
          },
          {
            src: "apple-touch-icon-iphone-60x60.png",
            sizes: "60x60",
            type: "image/png"
          },
          {
            src: "apple-touch-icon-ipad-76x76.png",
            sizes: "76x76",
            type: "image/png"
          },
          {
            src: "apple-touch-icon-iphone-retina-120x120.png",
            sizes: "120x120",
            type: "image/png"
          },
          {
            src: "apple-touch-icon-ipad-retina-152x152.png",
            sizes: "152x152",
            type: "image/png"
          }
        ],
        start_url: "/",
        scope: "/",
        display: "standalone"
      },
      workbox: {
        // looks inside "dist" folder
        globPatterns: ["**/*.{js,css,html,svg,woff2}"],
        sourcemap: true,
        navigateFallbackDenylist: [/^\/api/]
      }
    }),
    ],
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
