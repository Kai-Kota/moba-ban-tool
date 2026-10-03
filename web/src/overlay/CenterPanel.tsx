import { logoUrl } from "../shared/api";
import type { DraftState, Side, Step } from "../shared/types";

// 中央: フェーズ表示、両チームのロゴとシリーズスコア、BO表記、パッチ。
export default function CenterPanel({ state, order }: { state: DraftState; order: Step[] }) {
  const step = state.phase === "drafting" ? order[state.currentStep] : undefined;
  const phaseText = step
    ? (step.label ?? `${step.type === "ban" ? "BAN" : "PICK"} PHASE`)
    : "DRAFT COMPLETE";
  const gameNo = Math.min(state.blue.seriesWins + state.red.seriesWins + 1, state.bestOf);

  return (
    <div className="center">
      <div className={`phase ${step?.team ?? "done"}`}>
        <span className="phase-text">{phaseText}</span>
        {step && (
          <span className="phase-turn">
            {state[step.team].teamName} {step.type === "ban" ? "BAN" : "PICK"}
          </span>
        )}
      </div>

      <div className="score">
        <TeamMark side="blue" state={state} />
        <div className="score-nums">
          <span className="score-blue">{state.blue.seriesWins}</span>
          <span className="score-sep">-</span>
          <span className="score-red">{state.red.seriesWins}</span>
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
      <div className="mark-logo">
        {t.logo ? (
          <img src={logoUrl(t.logo)} alt="" />
        ) : (
          <span className="mark-initial">{t.teamName.slice(0, 1)}</span>
        )}
      </div>
      {state.bestOf > 1 && (
        <div className="pips">
          {Array.from({ length: winsNeeded }, (_, i) => (
            <span key={i} className={i < t.seriesWins ? "pip won" : "pip"} />
          ))}
        </div>
      )}
    </div>
  );
}
