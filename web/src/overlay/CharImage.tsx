import { useState } from "react";
import { characterImageUrl } from "../shared/api";
import type { Character } from "../shared/types";

// ページを開いた直後(OBSでソースを読み込んだ直後や再接続直後)に、
// すでに埋まっている枠が一斉にアニメーションしないようにするためのフラグ。
// 最初の表示が終わってから true になり、それ以降にマウントされた画像だけが入場演出を行う。
let animationsReady = false;
export function enableAnimationsSoon() {
  if (!animationsReady) window.setTimeout(() => (animationsReady = true), 300);
}

// キャラ画像。key に charId を指定して使うと、キャラが変わるたびに作り直されて入場演出が走る。
export default function CharImage({ char, className }: { char: Character; className: string }) {
  const [animate] = useState(() => animationsReady);
  return (
    <img
      className={`${className}${animate ? " enter" : ""}`}
      src={characterImageUrl(char)}
      alt=""
      draggable={false}
    />
  );
}

// 入場演出用の装飾(フラッシュ、斜線など)。CharImage と同じ条件で .enter が付く。
export function EnterFx({ className }: { className: string }) {
  const [animate] = useState(() => animationsReady);
  return <div className={`${className}${animate ? " enter" : ""}`} />;
}
