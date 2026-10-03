import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./overlay.css";

// OBS のブラウザソースには window.obsstudio がある。
// それ以外(通常のブラウザで確認するとき)は背景を付けて画面に収まるよう縮小する。
// ?preview=0 / ?preview=1 で強制的に切り替えられる。
const param = new URLSearchParams(location.search).get("preview");
const preview = param !== null ? param !== "0" : !("obsstudio" in window);
if (preview) document.body.classList.add("preview");

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App preview={preview} />
  </StrictMode>,
);
