import { logoUrl } from "../shared/api";
import type { DraftState, Side, Step } from "../shared/types";

// 中央: フェーズ表示、両チームの勝利数(丸)・ロゴ・チーム名、パッチとBO表記。
export default function CenterPanel({ state, order }: { state: DraftState; order: Step[] }) {
  const step = state.phase === "drafting" ? order[state.currentStep] : undefined;
  const phaseText = step
    ? (step.label ?? `${step.type === "ban" ? "BAN" : "PICK"} PHASE`)
    : "DRAFT COMPLETE";
  const gameNo = Math.min(state.blue.seriesWins + state.red.seriesWins + 1, state.bestOf);

  return (
    <div className="center">
      <div className="phase-text">{phaseText}</div>

      <div className="score">
        <TeamMark side="blue" state={state} />
        <div className={`turn-mark ${step?.team ?? "done"}`}>
          <span className="turn-arrow" />
          <span className="turn-type">{step ? step.type.toUpperCase() : ""}</span>
        </div>
        <TeamMark side="red" state={state} />
      </div>

      <div className="series">
        <span>GAME {gameNo}</span>
        <span className="dot">•</span>
        <span>BO{state.bestOf}</span>
        {state.patch && (
          <>
            <span className="dot">•</span>
            <span>PATCH {state.patch}</span>
          </>
        )}
      </div>
    </div>
  );
}

function TeamMark({ side, state }: { side: Side; state: DraftState }) {
  const t = state[side];
  const winsNeeded = Math.floor(state.bestOf / 2) + 1;
  return (
    <div className={`mark ${side}`}>
      <div className="pips">
        {state.bestOf > 1 &&
          Array.from({ length: winsNeeded }, (_, i) => (
            <span key={i} className={i < t.seriesWins ? "pip won" : "pip"} />
          ))}
      </div>
      <div className="mark-logo">
        {t.logo ? (
          <img src={logoUrl(t.logo)} alt="" />
        ) : (
          <span className="mark-initial">{t.teamName.slice(0, 1)}</span>
        )}
      </div>
      <div className="mark-name">{t.teamName}</div>
    </div>
  );
}
