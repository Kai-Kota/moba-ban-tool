import { characterImageUrl, logoUrl } from "../shared/api";
import type { Character, DraftState, Side, Step, StepType } from "../shared/types";

interface Props {
  side: Side;
  state: DraftState;
  order: Step[];
  slots: number[];
  chars: Map<string, Character>;
}

// 片側チームのBAN/PICK枠の一覧。次に埋まる枠を強調し、選択中のキャラを薄く表示する。
export default function DraftBoard({ side, state, order, slots, chars }: Props) {
  const team = state[side];
  const step = state.phase === "drafting" ? order[state.currentStep] : undefined;
  const activeSlot = (type: StepType) =>
    step && step.team === side && step.type === type ? slots[state.currentStep] : -1;

  const cell = (charId: string, active: boolean, label?: string) => {
    const shown = charId || (active ? state.hover : "");
    const c = chars.get(shown);
    const cls = ["slot", active && "active", !charId && shown && "ghost"].filter(Boolean).join(" ");
    return (
      <div className={cls}>
        {c ? <img src={characterImageUrl(c)} alt="" /> : <div className="slot-empty" />}
        <div className="slot-text">
          {label !== undefined && <span className="slot-player">{label || "-"}</span>}
          <span className="slot-char">{c?.name ?? (shown || "")}</span>
        </div>
      </div>
    );
  };

  const banActive = activeSlot("ban");
  const pickActive = activeSlot("pick");
  return (
    <div className={`board ${side}`}>
      <div className="board-head">
        {team.logo && <img className="board-logo" src={logoUrl(team.logo)} alt="" />}
        <span className="board-name">{team.teamName}</span>
        <span className="board-wins">{team.seriesWins}</span>
      </div>
      <div className="board-bans">
        {team.bans.map((id, i) => (
          <div key={i}>{cell(id, i === banActive)}</div>
        ))}
      </div>
      <div className="board-picks">
        {team.picks.map((id, i) => (
          <div key={i}>{cell(id, i === pickActive, team.players[i])}</div>
        ))}
      </div>
    </div>
  );
}
