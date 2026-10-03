import type { Side } from "../shared/types";
import { useDraftSocket } from "../shared/useDraftSocket";

// マイルストーン1の簡易オーバーレイ。state をテキストで表示する。
export default function App() {
  const { state } = useDraftSocket();
  if (!state) return null;

  const team = (side: Side) => {
    const t = state[side];
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
    <div className="stage">
      {team("blue")}
      <div className="center">
        <p>
          Step {state.currentStep + 1} / {state.phase}
        </p>
        {state.hover && <p className="hover">選択中: {state.hover}</p>}
      </div>
      {team("red")}
    </div>
  );
}
