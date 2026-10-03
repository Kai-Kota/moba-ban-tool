import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

const GO_SERVER = "http://127.0.0.1:8080";

// dev server でも本番(Goサーバー)と同じく /control, /overlay でアクセスできるようにする
function pageRoutes(): Plugin {
  const pages: Record<string, string> = {
    "/control": "/control.html",
    "/overlay": "/overlay.html",
  };
  return {
    name: "page-routes",
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        const [path, query] = (req.url ?? "").split("?");
        const page = pages[path.replace(/\/$/, "")];
        if (page) req.url = page + (query ? `?${query}` : "");
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), pageRoutes()],
  server: {
    proxy: {
      "/ws": { target: GO_SERVER, ws: true },
      "/api": GO_SERVER,
      "/assets": GO_SERVER,
    },
  },
  build: {
    // OBS 30 のブラウザソース/ドックは Chromium 103 ベース。
    // 既定のままだと @media (max-width: ...) が範囲構文 (width<=...) に変換され、OBS で効かなくなる
    target: "chrome100",
    cssTarget: "chrome100",
    // Vite の既定の出力先 assets/ は Go 側の /assets/(data/ の画像)と衝突するため変える
    assetsDir: "static",
    rollupOptions: {
      input: {
        control: import.meta.dirname + "/control.html",
        overlay: import.meta.dirname + "/overlay.html",
      },
    },
  },
});
