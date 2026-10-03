import { useEffect, useState } from "react";
import type { Side } from "../shared/types";
import { useDraftSocket } from "../shared/useDraftSocket";

const STAGE_W = 1920;
const STAGE_H = 1080;

// プレビュー時に 1920x1080 をウィンドウに収める倍率
function useFitScale(enabled: boolean): number {
  const calc = () => Math.min(innerWidth / STAGE_W, innerHeight / STAGE_H);
  const [scale, setScale] = useState(enabled ? calc : 1);
  useEffect(() => {
    if (!enabled) return;
    const onResize = () => setScale(calc());
    addEventListener("resize", onResize);
    return () => removeEventListener("resize", onResize);
  }, [enabled]);
  return enabled ? scale : 1;
}

// マイルストーン1の簡易オーバーレイ。state をテキストで表示する。
export default function App({ preview }: { preview: boolean }) {
  const { state, connected } = useDraftSocket();
  const scale = useFitScale(preview);

  const team = (side: Side) => {
    const t = state![side];
    return (
      <div className={`team ${side}`}>
        <h1>
          {t.teamName} ({t.seriesWins})
        </h1>
        <p>BAN: {t.bans.map((b) => b || "-").join(" / ")}</p>
        <p>PICK: {t.picks.map((p) => p || "-").join(" / ")}</p>
      </div>
    );
  };

  return (
    <div className="stage" style={{ transform: `scale(${scale})` }}>
      {preview && !connected && <div className="preview-status">サーバー接続待ち...</div>}
      {state && (
        <>
          {team("blue")}
          <div className="center">
            <p>
              Step {state.currentStep + 1} / {state.phase}
            </p>
            {state.hover && <p className="hover">選択中: {state.hover}</p>}
          </div>
          {team("red")}
        </>
      )}
    </div>
  );
}
