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

// 手番ごとのタイマーの秒数(既定30秒)。?timer=0 で非表示
const timerParam = Number(new URLSearchParams(location.search).get("timer") ?? "30");
const timerSec = Number.isFinite(timerParam) && timerParam > 0 ? timerParam : 0;

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App preview={preview} timerSec={timerSec} />
  </StrictMode>,
);
